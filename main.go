package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Do these structs need to be public
type Node struct {
	id          string
	state       string
	currentTerm int
	votedFor    string
	commitIndex int
	//log
}

type VoteRequest struct {
	Term         int    `json:"term"`
	CandidateId  string `json:"candidateId"`
	LastLogIndex int    `json:"lastLogIndex"`
	LastLogTerm  int    `json:"lastLogTerm"`
}

type VoteResponse struct {
	Term        int  `json:"term"`
	VoteGranted bool `json:"voteGranted"`
}

func newNode(id string) Node {
	return Node{
		currentTerm: 0,
		votedFor:    "",
		state:       "Follower",
		id:          id,
	}
}

func main() {
	heartbeat := make(chan string, 1)
	var wg sync.WaitGroup

	var ports [3]string
	ports[0] = "8080"
	ports[1] = "8081"
	ports[2] = "8082"

	http.HandleFunc("/appendEntries", appendEntries(heartbeat))

	portIndex := 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, port := range ports {
			fmt.Printf("Connecting to localhost:%s...\n", port)
			server(port)
			portIndex++
		}
	}()

	time.Sleep(1 * time.Second)

	fmt.Println("PortIndex: ", portIndex)
	node := newNode(ports[portIndex])
	http.HandleFunc("/requestVote", requestVote(node))
	client := &http.Client{}
	//Think of a better way to do this
	votesRecieved := 0

	//Follower loop
	for {
		if node.state == "Follower" {
			select {
			case <-heartbeat:
				fmt.Println("Following")
			case <-time.After(time.Second * 3):
				//Increment current term and transition to candidate state
				node.state = "Candidate"
				node.currentTerm++
				fmt.Println("I should be leader")
				//Send Request Vote to all servers in the cluster
				vReq := VoteRequest{
					Term:        node.currentTerm,
					CandidateId: node.id,
				}
				data, err := json.Marshal(vReq)
				if err != nil {
					fmt.Println("Error while martialing vote request")
				}
				//Send the vote request to all servers
				//How do I know about all servers?
				for _, port := range ports {
					if port != node.id {
						url := "http://localhost:" + port + "/requestVote"

						//jsonStr := []byte(`{"key":"value"}`)
						req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
						if err != nil {
							//TO DO - make this catch more specific
							fmt.Println("We have an error: ", err)
						}
						res, err := client.Do(req)
						if err != nil {
							fmt.Println("Error encountered when doing post request:", err)
						}
						//Process the response to see if am new leader
						defer res.Body.Close()
						body, err := io.ReadAll(res.Body)
						if err != nil {
							fmt.Println("Could not read the response body")
						}
						vRes := VoteResponse{}
						err = json.Unmarshal(body, &vRes)
						if err != nil {
							fmt.Println("Could not unmarshall requestVote json")
						}
						fmt.Println(vRes)
						if vRes.VoteGranted {
							votesRecieved++
						}
					}
				}
				if votesRecieved >= 2 {
					node.state = "Leader"
					sendHeartbeat(node, ports)
					fmt.Println("I'm the leader")
				}
			}
		}
		if node.state == "Candidate" {
			select {
			case <-heartbeat:
				node.state = "Follower"
			case <-time.After(time.Second * 3):
				node.currentTerm++
				fmt.Println("I should still be leader")
				//Here we should start another election
			}
		}
		if node.state == "Leader" {
			sendHeartbeat(node, ports)
			time.Sleep(2 * time.Second)
		}
	}
}

func server(port string) {
	err := http.ListenAndServe(":"+port, nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("server closed")
		//Is there a specific error for port alreadyin use?
	} else if err != nil {
		fmt.Printf("Port %s is in use, trying next one\n", port)
	}
}

func sendHeartbeat(node Node, ports [3]string) {
	for _, port := range ports {
		if port != node.id {
			url := "http://localhost:" + port + "/appendEntries"
			jsonStr := []byte(`{"key":"value"}`)
			client := &http.Client{}
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
			if err != nil {
				//TO DO - make this catch more specific
				fmt.Println("We have an error: ", err)
			}
			_, err = client.Do(req)
			if err != nil {
				fmt.Println("Error encountered when sending heart beat:", err)
			}
		}
	}
}

func appendEntries(heart chan string) http.HandlerFunc {
	// A handler fucntion for AppendEntries RPCs
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("Still alive")
		heart <- "PING"
	}
}

func requestVote(node Node) http.HandlerFunc {
	// A handler fucntion for RequestVote RPCs
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			fmt.Println("Could not read the response body")
		}
		fmt.Println(string(body))
		vReq := VoteRequest{}
		err = json.Unmarshal(body, &vReq)
		if err != nil {
			fmt.Println("Could not unmarshall requestVote json")
		}
		fmt.Println(vReq)

		vRes := VoteResponse{}
		//Check here to make sure logs of candidate are up to date
		//Clean this up a bit
		if vReq.Term >= node.currentTerm && (node.votedFor == "" || node.votedFor == vReq.CandidateId) {
			vRes.VoteGranted = true
			vRes.Term = node.currentTerm
			node.votedFor = vReq.CandidateId
		} else {
			vRes.VoteGranted = false
			vRes.Term = node.currentTerm
		}
		data, err := json.Marshal(vRes)
		if err != nil {
			fmt.Println("Error marshaling json in requestVote")
		}
		w.Write(data)
	}
}

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

type Node struct {
	id          int
	state       string
	currentTerm int
	votedFor    int
	commitIndex int
	//log
}

type VoteRequest struct {
	Term         int `json:"term"`
	CandidateId  int `json:"candidateId"`
	LastLogIndex int `json:"lastLogIndex"`
	LastLogTerm  int `json:"lastLogTerm"`
}

// A handler fucntion for RequestVote RPCs
func requestVote(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Println("Could not read the response body")
	}
	fmt.Println(string(body))
	vReq := VoteRequest{}
	_ = json.Unmarshal(body, &vReq)
	fmt.Println(vReq)
}

func newNode(id int) Node {
	return Node{
		currentTerm: 0,
		votedFor:    -1,
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
	http.HandleFunc("/requestVote", requestVote)

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
	node := newNode(portIndex)
	client := &http.Client{}

	//Follower loop
	for {
		if node.state == "Follower" {
			select {
			case <-heartbeat:
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
				for i, port := range ports {

					if i != node.id {
						url := "http://localhost:" + port + "/requestVote"

						//jsonStr := []byte(`{"key":"value"}`)
						req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
						if err != nil {
							//TO DO - make this catch more specific
							fmt.Println("We have an error: ", err)
						}
						_, err = client.Do(req)
						if err != nil {
							fmt.Println("Error encountered when doing post request:", err)
						}
					}
				}
			}
		}
		if node.state == "Candidate" {
			select {
			case <-heartbeat:
			case <-time.After(time.Second * 3):
				node.currentTerm++
				fmt.Println("I should still be leader")
				//Send Request Vote to all servers in the cluster
				// req := VoteRequest{
				// 	Term:        node.currentTerm,
				// 	CandidateId: node.id,
				// }
			}
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

func appendEntries(heart chan string) http.HandlerFunc {
	// A handler fucntion for AppendEntries RPCs
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("Still alive")
		heart <- "PING"
	}
}

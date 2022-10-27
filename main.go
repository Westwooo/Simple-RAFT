package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	clearLine = "\033[2K"
	upLine    = "\033[A"
)

//How could I use a RAFT interface to make this neater?
//Create HTTP manager that has http,client, ports maybe?

type RAFT interface {
	voteForMe(VoteRequest, string) VoteResponse //Send a vote to the node identified by the string
	setConns()                                  //Set the connections of a node appropriately
	runServer(Node)                             //
}

//How to implement the server on the node?

// Do these structs need to be public
type Node struct {
	id            string
	state         string
	currentTerm   int
	votedFor      string
	commitIndex   int
	svrs          int
	votesRecieved int
	heart         chan int
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

type AppendRequest struct {
	Term    int      `json:"term"`
	Entries []string `json:"entries"`
}

type AppendResponse struct {
	Term    int  `json:"term"`
	Success bool `json:"success"`
}

func newNode(id string) Node {
	return Node{
		currentTerm: 0,
		votedFor:    "",
		state:       "Follower",
		id:          id,
		svrs:        0,
	}
}

func main() {
	var wg sync.WaitGroup
	rand.Seed(time.Now().UnixNano())

	//Node has list of connections
	var ports [4]string
	ports[0] = "8080"
	ports[1] = "8081"
	ports[2] = "8082"
	ports[3] = "8083"

	portIndex := 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, port := range ports {
			//fmt.Printf("Connecting to localhost:%s...\n", port)
			server(port)
			portIndex++
		}
	}()

	time.Sleep(1 * time.Second)

	fmt.Printf("Listening on port: %s\n", ports[portIndex])

	node := newNode(ports[portIndex])
	node.heart = make(chan int, 1)
	fmt.Print("\n\n")
	displayNode(node)

	//TO DO: make client part of node
	client := &http.Client{}
	http.HandleFunc("/requestVote", requestVote(&node))
	http.HandleFunc("/appendEntries", appendEntries(node.heart))

	electionTime := time.Duration(0)

	for {
		displayNode(node)
		if node.state == "Follower" {
			select {
			case lTerm := <-node.heart:
				node.currentTerm = lTerm
			case <-time.After((time.Second * 3)):
				//Follower -> Candidate needs this
				fmt.Println("")
				node.state = "Candidate"
				electionTime = 0
			}
		}
		if node.state == "Candidate" {
			select {
			case lTerm := <-node.heart:
				//Candidate -> Follower needs this
				fmt.Print(clearLine, upLine, clearLine)
				node.state = "Follower"
				node.currentTerm = lTerm
			case <-time.After(electionTime):
				startElection(&node, ports, client)
				if node.votesRecieved > node.svrs/2 {
					node.state = "Leader"
					sendHeartbeat(&node, ports)
				} else {
					electionTime = time.Duration(rand.Intn(1000000000) * 3)
				}
			}
		}
		if node.state == "Leader" {
			displayNode(node)
			sendHeartbeat(&node, ports)
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
		//fmt.Printf("Port %s is in use, trying next one\n", port)
	}
}

func sendHeartbeat(n *Node, ports [4]string) {
	n.svrs = 1
	for _, port := range ports {
		if port != n.id {
			url := "http://localhost:" + port + "/appendEntries"

			aReq := AppendRequest{
				Term:    n.currentTerm,
				Entries: nil,
			}
			data, err := json.Marshal(aReq)
			client := &http.Client{}
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
			if err != nil {
				//TO DO - make this catch more specific
				fmt.Println("We have an error: ", err)
			}
			_, err = client.Do(req)
			if err != nil {
				//TO DO: is tjere a reason to handle this?
				//fmt.Println("Error encountered when sending heart beat:", err)
			} else {
				n.svrs++
			}
		}
	}
}

func appendEntries(heart chan int) http.HandlerFunc {
	// A handler fucntion for AppendEntries RPCs
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			fmt.Println("Could not read the response body")
		}
		aReq := AppendRequest{}
		err = json.Unmarshal(body, &aReq)
		if err != nil {
			//fmt.Println("Could not unmarshall requestVote json")
		} else {
			if aReq.Entries == nil {
				heart <- aReq.Term
			}
			//else contains log entries to be added
		}
	}
}

//We want http handler function to

func requestVote(node *Node) http.HandlerFunc {
	// A handler fucntion for RequestVote RPCs
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			fmt.Println("Could not read the response body")
		}
		vReq := VoteRequest{}
		err = json.Unmarshal(body, &vReq)
		if err != nil {
			//fmt.Println("Could not unmarshall requestVote json")
		} else {
			vRes := node.response(vReq)
			data, err := json.Marshal(vRes)
			if err != nil {
				//fmt.Println("Error marshaling json in requestVote")
			}
			w.Write(data)
		}
	}
}

func (n *Node) response(vReq VoteRequest) VoteResponse {
	vRes := VoteResponse{}
	if vReq.Term > n.currentTerm && (n.votedFor == "" || n.votedFor == vReq.CandidateId) {
		vRes.VoteGranted = true
		n.currentTerm = vReq.Term
		n.votedFor = vReq.CandidateId
	} else {
		vRes.VoteGranted = false
		vRes.Term = n.currentTerm
	}
	return vRes
}

func startElection(n *Node, ps [4]string, c *http.Client) {
	//startElection(node, ports, client)
	//Increment current term and transition to candidate state
	n.state = "Candidate"
	n.currentTerm++
	displayNode(*n)

	//Send Request Vote to all servers in the cluster
	vReq := VoteRequest{
		Term:        n.currentTerm,
		CandidateId: n.id,
	}

	//Node knows it is in cluster, and votes for itself
	n.svrs = 1
	n.votesRecieved = 1

	//Paralellize this
	for _, port := range ps {
		if port != n.id {
			vRes := n.voteForMe(vReq, port, c)
			//fmt.Println(vRes)
			if vRes.VoteGranted {
				n.votesRecieved++
			}
		}
	}
	displayNode(*n)
}

// voteForMe takes a voteRequest and resturns a voteResponse,
// Node doesn't need to worry how it gets there
func (n *Node) voteForMe(vReq VoteRequest, con string, c *http.Client) VoteResponse {
	data, err := json.Marshal(vReq)
	if err != nil {
		fmt.Println("Error while marshaling vote request")
	}
	url := "http://localhost:" + con + "/requestVote"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("error creating HTTP request: ", err)
	}
	res, err := c.Do(req)
	if err != nil {
		//Error will occur when sending vote to servers that are down
	} else {
		n.svrs++
		//Process the response to see if am new leader
		//Create inline function to trigger defern earlier
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			//fmt.Println("Could not read the response body")
		}
		vRes := VoteResponse{}

		err = json.Unmarshal(body, &vRes)
		if err != nil {
			//fmt.Println("Could not unmarshall requestVote json")
		}
		return vRes
	}
	return VoteResponse{VoteGranted: false}
}

func displayNode(node Node) {
	if node.state == "Leader" {
		fmt.Print(clearLine, upLine, clearLine, upLine, clearLine, upLine, clearLine)
		fmt.Println("Status: ", node.state)
		fmt.Println("Term: ", node.currentTerm)
		fmt.Println("Servers in cluster :", node.svrs)
	} else if node.state == "Follower" {
		fmt.Print(clearLine, upLine, clearLine, upLine, clearLine)
		fmt.Println("Status: ", node.state)
		fmt.Println("Term: ", node.currentTerm)
	} else if node.state == "Candidate" {
		fmt.Print(clearLine, upLine, clearLine, upLine, clearLine, upLine, clearLine)
		fmt.Println("Status: ", node.state)
		fmt.Println("Term: ", node.currentTerm)
		fmt.Println("Votes Recieved: ", node.votesRecieved)
	}
}

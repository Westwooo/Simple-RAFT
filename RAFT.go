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
	voteForMe(VoteRequest, string) VoteResponse //Send a voteRequest to the node identified by the string
	setConns()                                  //Set the connections of a node appropriately
	runServer(Node)                             //Run the server listening in the appropriate way
}

//How to implement the server on the node?

// Do these structs need to be public
type Node struct {
	heart         chan int
	id            string
	state         string
	votedFor      string
	conns         []string
	currentTerm   int
	commitIndex   int
	svrs          int
	votesRecieved int
	wg            sync.WaitGroup
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

func newNode() Node {
	return Node{
		currentTerm: 0,
		votedFor:    "",
		state:       "Follower",
		svrs:        0,
		heart:       make(chan int, 1),
	}
}

func (n *Node) runServer() {
	portIndex := 0
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		for _, port := range n.conns {
			err := http.ListenAndServe(":"+port, nil)
			if errors.Is(err, http.ErrServerClosed) {
				fmt.Println("server closed")
			} else if err != nil {
				//This port is in use, try the next one
			}
			portIndex++
		}
	}()
	time.Sleep(1 * time.Second)
	n.id = n.conns[portIndex]
}

func main() {

	//Seed random generator for the generated times later
	rand.Seed(time.Now().UnixNano())

	//Create a new node, set connections and run the server
	node := newNode()
	node.setConns()
	node.runServer()

	fmt.Printf("Listening on port: %s\n", node.id)

	//node.id = node.conns[portIndex]

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
				startElection(&node, client)
				if node.votesRecieved > node.svrs/2 || (node.votesRecieved >= node.svrs/2 && node.svrs%2 == 0) {
					node.state = "Leader"
					sendHeartbeat(&node)
				} else {
					electionTime = time.Duration(rand.Intn(1000000000) * 3)
				}
			}
		}
		if node.state == "Leader" {
			displayNode(node)
			select {
			case lTerm := <-node.heart:
				if lTerm > node.currentTerm {
					node.state = "Follower"
				} else if lTerm == node.currentTerm {
					node.state = "Candidate"
					electionTime = time.Duration(rand.Intn(1000000000) * 3)
				}
			case <-time.After(2 * time.Second):
				sendHeartbeat(&node)
			}
		}
	}
}

func sendHeartbeat(n *Node) {
	n.svrs = 1
	for _, port := range n.conns {
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

func (n *Node) setConns() {
	var conns []string
	conns = append(conns, "8080", "8081", "8082", "8083", "8084")
	n.conns = conns
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

// We want http handler function to
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

func startElection(n *Node, c *http.Client) {
	//startElection(node, ports, client)
	//Increment current term and transition to candidate state
	n.state = "Candidate"
	n.votedFor = ""
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
	for _, port := range n.conns {
		if port != n.id {
			vRes := n.voteForMe(vReq, port, c)
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

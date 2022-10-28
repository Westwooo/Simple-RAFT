package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	clearLine = "\033[2K"
	upLine    = "\033[A"
)

type RAFT interface {
	requestVotes(VoteRequest, string) VoteResponse      //Send a voteRequest to the node identified by the string
	appendEntries(AppendRequest, string) AppendResponse //Send an appendRequest to the node identified by the string
	runServer() error                                   //Run server and set node.id as endpoint used
	print()                                             //Print the node information in a nice way

	Heart() chan int //Return the channel used as the nodes heart

	State() string
	SetState(string)

	Term() int
	SetTerm(int)

	Conns() []string

	Svrs() int
	SetSvrs(int)

	Votes() int
	SetVotes(int)

	SetVotedFor(string)

	Id() string
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

func runRaft(node RAFT) {
	//Seed random generator for the generated times later
	rand.Seed(time.Now().UnixNano())

	//Run the server as appropriate
	err := node.runServer()
	if err != nil {
		fmt.Print(err)
		return
	}

	heart := node.Heart()

	electionTime := time.Duration(0)
	for {
		node.print()
		if node.State() == "Follower" {
			select {
			case lTerm := <-heart:
				node.SetTerm(lTerm)
			case <-time.After((time.Second * 3)):
				//Follower -> Candidate needs this
				fmt.Println("")
				node.SetState("Candidate")
				electionTime = time.Duration(rand.Intn(1000000000) * 3)
			}
		}
		if node.State() == "Candidate" {
			node.print()
			select {
			case lTerm := <-heart:
				//Vulnerable to malicious agent sending ping with inflated term
				fmt.Print(clearLine, upLine, clearLine)
				node.SetState("Follower")
				node.SetTerm(lTerm)
			case <-time.After(electionTime):
				startElection(node)
				if node.Votes() > node.Svrs()/2 || (node.Votes() >= node.Svrs()/2 && node.Svrs()%2 == 0) {
					node.SetState("Leader")
				} else {
					electionTime = time.Duration(rand.Intn(1000000000) * 3)
				}
			}
		}
		if node.State() == "Leader" {
			node.SetSvrs(1)
			for _, c := range node.Conns() {
				if c != node.Id() {
					aReq := AppendRequest{
						Term:    node.Term(),
						Entries: nil,
					}
					aRes := node.appendEntries(aReq, c)

					if aRes.Term != -1 {
						node.SetSvrs(node.Svrs() + 1)
					}
				}
			}
			node.print()

			select {
			case lTerm := <-heart:
				if lTerm > node.Term() {
					node.SetState("Follower")
				} else if lTerm == node.Term() {
					node.SetState("Candidate")
					electionTime = time.Duration(rand.Intn(1000000000) * 3)
				}
			case <-time.After(2 * time.Second):
			}
		}
	}
}

func startElection(node RAFT) {
	//startElection(node, ports, client)
	//Increment current term and transition to candidate state
	node.SetState("Candidate")
	node.SetVotedFor("")
	node.SetTerm(node.Term() + 1)
	node.print()

	//Send Request Vote to all servers in the cluster
	vReq := VoteRequest{
		Term:        node.Term(),
		CandidateId: node.Id(),
	}

	//Node knows it is in cluster, and votes for itself
	node.SetSvrs(1)
	node.SetVotes(1)

	//Paralellize this
	for _, c := range node.Conns() {
		if c != node.Id() {
			vRes := node.requestVotes(vReq, c)
			if vRes.VoteGranted {
				node.SetVotes(node.Votes() + 1)
			}
		}
	}
	node.print()
}

// Node contructor takes endpoints as strings, length is max number of nodes
func newNode(client interface{}, conns ...string) Node {
	return Node{
		currentTerm: 0,
		votedFor:    "",
		state:       "Follower",
		svrs:        0,
		heart:       make(chan int, 1),
		conns:       conns,
		client:      client,
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

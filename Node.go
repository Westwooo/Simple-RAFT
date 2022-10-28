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

// Underlying struct for HTTP implementation of RAFT consensus algorithm
type Node struct {
	heart         chan int
	id            string
	state         string
	votedFor      string
	conns         []string
	currentTerm   int
	svrs          int
	votesRecieved int
	client        interface{}
}

func (n *Node) appendEntries(aReq AppendRequest, c string) AppendResponse {
	url := "http://localhost:" + c + "/recieveEntries"
	data, err := json.Marshal(aReq)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		//TO DO - make this catch more specific
		fmt.Println("We have an error: ", err)
	}
	_, err = n.client.(*http.Client).Do(req)
	if err != nil {
		return AppendResponse{
			Term: -1,
		}
	} else {
		return AppendResponse{
			Term: 1,
		}
	}
}

func (n *Node) requestVotes(vReq VoteRequest, con string) VoteResponse {
	cli := n.client.(*http.Client)
	data, err := json.Marshal(vReq)
	if err != nil {
		fmt.Println("Error while marshaling vote request")
	}
	url := "http://localhost:" + con + "/requestVote"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("error creating HTTP request: ", err)
	}
	res, err := cli.Do(req)
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

func (n *Node) runServer() error {
	portIndex := 0

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
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
	time.Sleep(1 * time.Millisecond)
	if portIndex == len(n.conns) {
		return errors.New("all endpoints are in use")
	}
	n.id = n.conns[portIndex]
	return nil
}

func (n *Node) Heart() chan int {
	return n.heart
}

func (n *Node) SetState(s string) {
	n.state = s
}

func (n *Node) State() string {
	return n.state
}

func (n *Node) Term() int {
	return n.currentTerm
}

func (n *Node) SetTerm(t int) {
	n.currentTerm = t
}

func (n *Node) Conns() []string {
	return n.conns
}

func (n *Node) Svrs() int {
	return n.svrs
}

func (n *Node) SetSvrs(s int) {
	n.svrs = s
}

func (n *Node) Votes() int {
	return n.votesRecieved
}

func (n *Node) SetVotes(v int) {
	n.votesRecieved = v
}

func (n *Node) Id() string {
	return n.id
}

func (n *Node) SetVotedFor(voted string) {
	n.votedFor = voted
}

func (n *Node) print() {
	if n.state == "Leader" {
		fmt.Print(clearLine, upLine, clearLine, upLine, clearLine, upLine, clearLine)
		fmt.Println("Status: ", n.state)
		fmt.Println("Term: ", n.currentTerm)
		fmt.Println("Servers in cluster :", n.svrs)
	} else if n.state == "Follower" {
		fmt.Print(clearLine, upLine, clearLine, upLine, clearLine)
		fmt.Println("Status: ", n.state)
		fmt.Println("Term: ", n.currentTerm)
	} else if n.state == "Candidate" {
		fmt.Print(clearLine, upLine, clearLine, upLine, clearLine, upLine, clearLine)
		fmt.Println("Status: ", n.state)
		fmt.Println("Term: ", n.currentTerm)
		fmt.Println("Votes Recieved: ", n.votesRecieved)
	}
}

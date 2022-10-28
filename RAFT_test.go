package main

import (
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestResponse(t *testing.T) {
	n := newNode(&http.Client{}, "8080", "8081", "8082", "8083", "8084")
	vReq := VoteRequest{
		Term:        1,
		CandidateId: "name",
	}
	//vReq.Term > n.currentTerm && n.votedFor == ""
	vRes := n.response(vReq)
	if !vRes.VoteGranted || vRes.Term != 0 {
		t.Errorf("response was incorrect, got [%v,%v], wanted [true,0]", vRes.VoteGranted, vRes.Term)
	}
	if n.currentTerm != vReq.Term || n.votedFor != vReq.CandidateId {
		t.Errorf("response was incorrect, got [%v,%v], wanted [1,name]", n.currentTerm, n.votedFor)
	}

	//vReq.Term > n.currentTerm && n.votedFor == vReq.candidateId
	n.votedFor = vReq.CandidateId
	n.currentTerm = 0
	vRes = n.response(vReq)
	if !vRes.VoteGranted || vRes.Term != 0 {
		t.Errorf("response was incorrect, got [%v,%v], wanted [true,0]", vRes.VoteGranted, vRes.Term)
	}
	if n.currentTerm != vReq.Term || n.votedFor != vReq.CandidateId {
		t.Errorf("response was incorrect, got [%v,%v], wanted [1,name]", n.currentTerm, n.votedFor)
	}

	//vReq.Term < n.currentTerm
	n.currentTerm = 2
	n.votedFor = ""
	vRes = n.response(vReq)
	if vRes.VoteGranted || vRes.Term != 2 {
		t.Errorf("response was incorrect, got [%v,%v], wanted [false,2]", vRes.VoteGranted, vRes.Term)
	}
	if n.currentTerm != 2 || n.votedFor != "" {
		t.Errorf("response was incorrect, got [%v,%v], wanted [2,``]", n.currentTerm, n.votedFor)
	}
}

func TestRunRaft(t *testing.T) {

	mNode := mockNode{
		currentTerm: 0,
		votedFor:    "",
		state:       "Follower",
		svrs:        0,
		heart:       make(chan int, 1),
	}
	ep := [3]string{"A", "B", "C"}
	for _, e := range ep {
		mNode.conns = append(mNode.conns, e)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runRaft(&mNode)
	}()

	time.Sleep(1 * time.Second)

	//Node should remain follower for up to three seconds
	if mNode.state != "Follower" {
		t.Errorf("response incorrect, got %s, expected Follower", mNode.state)
	}

	//Recieving a ping on heart should keep the node as a follower and update currentTerm
	mNode.heart <- 3
	time.Sleep(2010 * time.Millisecond)

	if mNode.state != "Follower" && mNode.currentTerm != 3 {
		t.Errorf("response incorrect, got (%s, %d), expected (Follower, 3)", mNode.state, mNode.currentTerm)
	}

	//After 3 seconds node should become a candidate
	time.Sleep(1010 * time.Millisecond)
	if mNode.state != "Candidate" && mNode.currentTerm != 3 {
		t.Errorf("response incorrect, got (%s, %d) expected (Candidate, 3)", mNode.state, mNode.currentTerm)
	}

	//Recieving a ping while candidate should retrun to follower
	mNode.heart <- 4
	time.Sleep(1 * time.Millisecond)
	if mNode.state != "Follower" && mNode.currentTerm != 4 {
		t.Errorf("response incorrect, got (%s, %d) expected (Follower, 4)", mNode.state, mNode.currentTerm)
	}

	time.Sleep(6 * time.Second)
	if mNode.state != "Leader" && mNode.currentTerm != 5 {
		t.Errorf("response incorrect, got (%s, %d) expected (Leader, 5)", mNode.state, mNode.currentTerm)
	}
}

func TestStartElection(t *testing.T) {
	mNode := mockNode{
		currentTerm: 0,
		votedFor:    "",
		state:       "Follower",
		svrs:        0,
		heart:       make(chan int, 1),
	}
	startElection(&mNode)
	time.Sleep(1 * time.Millisecond)
	if mNode.state != "Candidate" {
		t.Errorf("response incorrecet got %s, expected Candidate", mNode.state)
	}
	if mNode.currentTerm != 1 {
		t.Errorf("response incorrect got %d, expected 1", mNode.currentTerm)
	}
	if mNode.svrs != 1 {
		t.Errorf("response incorrect got %d, expected 1", mNode.svrs)
	}
	if mNode.votesRecieved != 1 {
		t.Errorf("response incorrect got %d, expected 1", mNode.votesRecieved)
	}

	ep := [3]string{"A", "B", "C"}
	for _, e := range ep {
		mNode.conns = append(mNode.conns, e)
	}
	mNode.conns =
}

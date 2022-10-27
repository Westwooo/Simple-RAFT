package main

import (
	"testing"
)

func TestResponse(t *testing.T) {
	n := newNode()
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

func TestSetConns(t *testing.T) {
	n := newNode()

	if len(n.conns) != 0 {
		t.Errorf("expecting length of 0, instead got: %d", len(n.conns))
	}
	n.setConns()
	testArray := [5]string{"8080", "8081", "8082", "8083", "8084"}
	for i, p := range testArray {
		if !(n.conns[i] == p) {
			t.Errorf("expected [8080 8081 8082 8083 8084]")
			t.Errorf("got: %v", n.conns)
		}
	}
}

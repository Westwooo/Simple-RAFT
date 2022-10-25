package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type Node struct {
	state       int
	currentTerm int
	votedFor    int
	commitIndex int
	//log
}

type VoteRequest struct {
	Term         int
	CandidateId  int
	LastLogIndex int
	LastLogTerm  int
}

// A very simple handler function
func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
	fmt.Println("hello")
	body, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Println("Error encountered opening request body")
	}
	fmt.Fprintf(w, string(body))
}

// A handler fucntion for RequestVote RPCs
func RequestVote(w http.ResponseWriter, req *http.Request) {

}

func newNode() Node {
	return Node{
		currentTerm: 0,
		votedFor:    -1,
	}
}

func main() {
	heartbeat := make(chan string, 1)
	var wg sync.WaitGroup
	// - Initialise a node and enter follower loop
	// - Follopwer loop:
	// 	- Listen to the port for a heartbeat from leader
	// 	- If no message recieved before election timeout
	// 		- Begin election
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/appendEntries", appendEntries(heartbeat))

	fmt.Println("Listening on localhost:8090")
	wg.Add(1)
	go func() {
		defer wg.Done()
		server()
	}()

	time.Sleep(2 * time.Second)

	//Follower loop
	for {
		select {
		case <-heartbeat:
		case <-time.After(time.Second * 3):
			//Increment current term and transition to candidate state
			fmt.Println("I should be leader")
			return
		}
	}
}

func server() {
	err := http.ListenAndServe(":8090", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("server closed")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func appendEntries(heart chan string) http.HandlerFunc {
	// A handler fucntion for AppendEntries RPCs
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("Still alive")
		heart <- "PING"
	}
}

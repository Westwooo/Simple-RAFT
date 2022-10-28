package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func main() {
	//Create a Node, with given clients and endpoints
	node := newNode(&http.Client{}, "8080", "8081", "8082", "8083", "8084")

	http.HandleFunc("/requestVote", requestVote(&node))
	http.HandleFunc("/recieveEntries", recieveEntries(node.heart))

	runRaft(&node)
}

func recieveEntries(heart chan int) http.HandlerFunc {
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

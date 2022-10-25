package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

// A very simple handler function
func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n")
	body, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Println("Error encountered opening request body")
	}
	fmt.Fprintf(w, string(body))
}

// A handler fucntion for AppendEntrie RPCs
func AppendEntries(w http.ResponseWriter, req *http.Request) {

}

// A handler fucntion for RequestVote RPCs
func RequestVote(w http.ResponseWriter, req *http.Request) {

}

func main() {
	http.HandleFunc("/hello", hello)
	fmt.Println("Listening on localhost:8090")
	server()
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

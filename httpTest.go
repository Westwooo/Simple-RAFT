package main

import (
	"bytes"
	"fmt"
	"net/http"
)

func main() {
	httpTest()
}

func httpTest() {
	url := "http://localhost:8090/hello"
	jsonStr := []byte(`{"title":"Buy cheese and bread for breakfast."}`)
	// resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		//TO DO - make this catch more specific
		fmt.Println("We have an error: ", err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error encountered when doing post request")
	}

	g_req, err := http.NewRequest("GET", url, nil)
	resp, err = client.Do(g_req)
	if err != nil {
		fmt.Println("Error encountered when doing get request")
	}
	fmt.Println("response Status:", resp)
}

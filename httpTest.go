package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

func main() {
	httpTest()
}

func httpTest() {
	url := "http://localhost:8090/appendEntries"
	jsonStr := []byte(`{"key":"value"}`)
	client := &http.Client{}

	for i := 0; i < 3; i++ {
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
		if err != nil {
			//TO DO - make this catch more specific
			fmt.Println("We have an error: ", err)
		}
		_, err = client.Do(req)
		if err != nil {
			fmt.Println("Error encountered when doing post request:", err)
		}
		time.Sleep(2 * time.Second)
	}

	// req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	// if err != nil {
	// 	//TO DO - make this catch more specific
	// 	fmt.Println("We have an error: ", err)
	// }
	// _, err = client.Do(req)
	// if err != nil {
	// 	fmt.Println("Error encountered when doing post request:", err)
	// }
	// time.Sleep(2 * time.Second)
	//}

	// req, err = http.NewRequest("POST", "http://localhost:8090/kill", bytes.NewBuffer(jsonStr))
	// _, err = client.Do(req)
	// if err != nil {
	// 	fmt.Println("Error encountered when doing post request:", err)
	// }
	// time.Sleep(1 * time.Second)

	//body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	fmt.Println("Error reading the response body")
	// }
	// fmt.Println("response:", string(body))
}

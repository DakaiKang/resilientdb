package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

var jsonMap map[string]interface{}

func call_verifier(payload []byte) {

	url := "http://152.70.114.32/write"
	// fmt.Println("payload:>", string(payload))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Connection", "close")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	// fmt.Println("response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header)
	// body, _ := ioutil.ReadAll(resp.Body)
	ioutil.ReadAll(resp.Body)
	// fmt.Println("response Body:", string(body))

}
func main() {
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 100000
	content, _ := ioutil.ReadFile("verifier_http_request_sample.json")
	json_string := string(content)
	json.Unmarshal([]byte(json_string), &jsonMap)

	for i := 0; i < 1000000; i++ {
		for j := 0; j < 3; j++ {
			jsonMap["sequenceNumber"] = strconv.Itoa(i)
			// fmt.Println("sent:>", i)
			verifier_payload, _ := json.Marshal(jsonMap)
			go call_verifier(verifier_payload)
		}
	}

}

package main

import (
	"bufio"
	"log"
	"net/http"
	"strings"
)

func main() {
	url := "http://localhost:8081/v1/chatserver/subscribe"

	payload := strings.NewReader("{\n\n}")

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("cache-control", "no-cache")
	//req.Header.Add("Postman-Token", "abd38440-4873-420b-bddb-1b4780522a53")

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	reader := bufio.NewReader(res.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			log.Print(err)
			break
		}
		log.Println(string(line))
	}
}

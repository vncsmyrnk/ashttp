package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/ashttp/internal"
)

func main() {
	flag.Parse()
	args := flag.Args()

	action, err := NewAction(args)
	if err != nil {
		log.Println(err)
	}

	request := action.Request()
	config := action.Config()

	req, err := request.ToHTTPRequest(config)
	if err != nil {
		log.Println(err)
	}

	response, err := internal.Execute(req)
	if err != nil {
		log.Println(err)
	}

	output, err := prettyResponse(response)
	if err != nil {
		log.Println(err)
	}

	fmt.Println(output)
}

func prettyResponse(resp []byte) (string, error) {
	var data any
	err := json.Unmarshal(resp, &data)
	if err != nil {
		return string(resp), nil
	}

	pretty, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return string(pretty), nil
}

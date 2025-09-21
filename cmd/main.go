package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/ashttp/internal/http"
)

var cliFormatExpected = "<URL-alias> <http-method> [path-components...] [--option value]"

func main() {
	flag.Parse()
	args := flag.Args()

	action, err := NewAction(args)
	if err != nil {
		switch {
		case errors.Is(err, errInvalidFormat):
			help()
		default:
			fatal("failed to build action from arguments: %v", err)
		}
	}

	request := action.Request()
	setting, err := action.Setting()
	if err != nil {
		fatal("failed to load setting: %v", err)
	}

	req, err := request.ToHTTPRequest(setting)
	if err != nil {
		fatal("failed to build request: %v", err)
	}

	response, err := http.Execute(req)
	if err != nil {
		fatal("failed to execute request: %v", err)
	}

	output, err := prettyResponse(response)
	if err != nil {
		fmt.Println(err)
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

func fatal(format string, v ...any) {
	fmt.Printf("[error] %s\n", fmt.Sprintf(format, v...))
	os.Exit(1)
}

func help() {
	fmt.Printf("usage: %s\n", cliFormatExpected)
	os.Exit(0)
}

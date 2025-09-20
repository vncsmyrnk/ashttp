package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Request struct {
	Path       string
	HTTPMethod string
	Headers    map[string]string
	Body       map[string]any
}

func (r Request) ToHTTPRequest(config Config) (*http.Request, error) {
	body, err := json.Marshal(r.Body)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s", config.Domain, r.Path)
	req, err := http.NewRequest(r.HTTPMethod, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range config.Headers {
		req.Header.Add(k, v)
	}

	for k, v := range r.Headers {
		req.Header.Add(k, v)
	}

	return req, nil
}

func Execute(req *http.Request) ([]byte, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ashttp/internal/config"
)

type Request struct {
	Path    string
	Headers map[string]string
	Body    map[string]any
}

func (r Request) ToHTTPRequest(config config.Setting) (*http.Request, error) {
	body, err := json.Marshal(r.Body)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/%s", config.Domain, r.Path)
	req, err := http.NewRequest(http.MethodGet, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range config.Headers {
		req.Header.Set(k, v)
	}

	for k, v := range r.Headers {
		req.Header.Set(k, v)
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

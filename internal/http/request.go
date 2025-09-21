package http

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ashttp/internal/config"
)

type Request struct {
	Path      string
	Method    string
	Headers   map[string]string
	Arguments map[string]string
}

func (r Request) ToHTTPRequest(setting config.Setting) (*http.Request, error) {
	req, err := r.buildHTTPRequest(setting)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range setting.Headers {
		req.Header.Set(k, v)
	}

	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

func (r Request) buildHTTPRequest(setting config.Setting) (*http.Request, error) {
	switch strings.ToUpper(r.Method) {
	case http.MethodGet, http.MethodDelete:
		queryString := QueryString(r.Arguments).ToURL()
		url := fmt.Sprintf("%s/%s", setting.URL, r.Path)
		if queryString != "" {
			url = fmt.Sprintf("%s?%s", url, queryString)
		}

		return http.NewRequest(r.Method, url, nil)
	default:
		return nil, fmt.Errorf("method not suported")
	}
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

package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ashttp/internal/config"
	"github.com/stretchr/testify/require"
)

func TestRequest_ToHTTPRequest(t *testing.T) {
	tests := []struct {
		name            string
		request         Request
		setting         config.Setting
		possibleURLs    []string
		expectedMethod  string
		expectedHeaders map[string]string
		expectError     bool
	}{
		{
			name: "basic GET request with no arguments",
			request: Request{
				Path:      "users",
				Method:    "get",
				Arguments: nil,
			},
			setting: config.Setting{
				URL: "https://api.example.com",
			},
			possibleURLs:   []string{"https://api.example.com/users"},
			expectedMethod: http.MethodGet,
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			name: "DELETE request with arguments",
			request: Request{
				Path:   "posts/123",
				Method: "delete",
				Arguments: map[string]string{
					"force": "true",
				},
			},
			setting: config.Setting{
				URL: "https://jsonplaceholder.typicode.com",
				Headers: map[string]string{
					"Authorization": "Bearer token123",
				},
			},
			possibleURLs:   []string{"https://jsonplaceholder.typicode.com/posts/123?force=true"},
			expectedMethod: http.MethodDelete,
			expectedHeaders: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token123",
			},
		},
		{
			name: "GET request with multiple arguments",
			request: Request{
				Path:   "api/v1/data",
				Method: "get",
				Arguments: map[string]string{
					"filter": "active",
					"limit":  "100",
				},
				Headers: map[string]string{
					"X-Custom-Header": "custom-value",
				},
			},
			setting: config.Setting{
				URL: "https://localhost:8080",
			},
			possibleURLs: []string{
				"https://localhost:8080/api/v1/data?filter=active&limit=100",
				"https://localhost:8080/api/v1/data?limit=100&filter=active",
			},
			expectedMethod: http.MethodGet,
			expectedHeaders: map[string]string{
				"Content-Type":    "application/json",
				"X-Custom-Header": "custom-value",
			},
		},
		{
			name: "request with unsupported method",
			request: Request{
				Path:   "users",
				Method: "post", // Not supported by buildHTTPRequest
			},
			setting: config.Setting{
				URL: "https://api.example.com",
			},
			expectError: true,
		},
		{
			name: "empty path",
			request: Request{
				Path:   "",
				Method: "get",
			},
			setting: config.Setting{
				URL: "https://api.example.com",
			},
			possibleURLs:   []string{"https://api.example.com/"},
			expectedMethod: http.MethodGet,
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			name: "headers override - request headers take precedence",
			request: Request{
				Path:   "override-test",
				Method: "get",
				Headers: map[string]string{
					"Authorization": "Bearer new-token",
					"Custom-Header": "request-value",
				},
			},
			setting: config.Setting{
				URL: "https://api.test.com",
				Headers: map[string]string{
					"Authorization":  "Bearer old-token",
					"Default-Header": "config-value",
				},
			},
			possibleURLs:   []string{"https://api.test.com/override-test"},
			expectedMethod: http.MethodGet,
			expectedHeaders: map[string]string{
				"Content-Type":   "application/json",
				"Authorization":  "Bearer new-token",
				"Custom-Header":  "request-value",
				"Default-Header": "config-value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := tt.request.ToHTTPRequest(tt.setting)

			if tt.expectError {
				require.Error(t, err, "ToHTTPRequest() should return an error")
				return
			}

			require.NoError(t, err, "ToHTTPRequest() should not return an error")
			require.Contains(t, tt.possibleURLs, req.URL.String(), "URL should match one of the possible valid URLs")
			require.Equal(t, tt.expectedMethod, req.Method, "HTTP method should match expected value")

			require.True(t, req.Body == nil || req.ContentLength == 0, "Request body should be empty")

			for expectedKey, expectedValue := range tt.expectedHeaders {
				actualValue := req.Header.Get(expectedKey)
				require.Equal(t, expectedValue, actualValue, "Header %s should match expected value", expectedKey)
			}

			for actualKey := range req.Header {
				_, expected := tt.expectedHeaders[actualKey]
				require.True(t, expected, "Unexpected header found: %s = %v", actualKey, req.Header.Get(actualKey))
			}
		})
	}
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		serverStatus   int
		serverHeaders  map[string]string
		expectedBody   string
		expectError    bool
	}{
		{
			name:           "successful JSON response",
			serverResponse: `{"message": "success", "data": {"id": 123}}`,
			serverStatus:   200,
			serverHeaders: map[string]string{
				"Content-Type": "application/json",
			},
			expectedBody: `{"message": "success", "data": {"id": 123}}`,
			expectError:  false,
		},
		{
			name:           "successful plain text response",
			serverResponse: "Hello, World!",
			serverStatus:   200,
			serverHeaders: map[string]string{
				"Content-Type": "text/plain",
			},
			expectedBody: "Hello, World!",
			expectError:  false,
		},
		{
			name:           "empty response body",
			serverResponse: "",
			serverStatus:   204,
			serverHeaders:  nil,
			expectedBody:   "",
			expectError:    false,
		},
		{
			name:           "error status but valid response",
			serverResponse: `{"error": "not found"}`,
			serverStatus:   404,
			serverHeaders: map[string]string{
				"Content-Type": "application/json",
			},
			expectedBody: `{"error": "not found"}`,
			expectError:  false,
		},
		{
			name:           "large response body",
			serverResponse: strings.Repeat("x", 10000),
			serverStatus:   200,
			serverHeaders:  nil,
			expectedBody:   strings.Repeat("x", 10000),
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				for key, value := range tt.serverHeaders {
					w.Header().Set(key, value)
				}

				w.WriteHeader(tt.serverStatus)

				fmt.Fprint(w, tt.serverResponse)
			}))
			defer server.Close()

			req, err := http.NewRequest("GET", server.URL, nil)
			require.NoError(t, err, "Should be able to create test request")

			body, err := Execute(req)

			if tt.expectError {
				require.Error(t, err, "Execute() should return an error")
				return
			}

			require.NoError(t, err, "Execute() should not return an error")
			require.Equal(t, tt.expectedBody, string(body), "Response body should match expected value")
		})
	}
}

func TestExecute_NetworkError(t *testing.T) {
	req, err := http.NewRequest("GET", "http://invalid-url-that-does-not-exist.test", nil)
	require.NoError(t, err, "Should be able to create test request")

	_, err = Execute(req)
	require.Error(t, err, "Execute() should return a network error")
}

func TestExecute_Integration(t *testing.T) {
	t.Run("full integration test", func(t *testing.T) {
		expectedResponse := `{"userId": 1, "name": "Test User"}`
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "application/json", r.Header.Get("Content-Type"), "Content-Type header should be application/json")
			require.Equal(t, "Bearer test-token", r.Header.Get("Authorization"), "Authorization header should be 'Bearer test-token'")
			require.Equal(t, "value", r.Header.Get("X-Custom"), "Custom header should be 'value'")
			require.Equal(t, "test", r.URL.Query().Get("query"), "URL query param 'query' should be 'test'")

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, expectedResponse)
		}))
		defer server.Close()

		ashttpRequest := Request{
			Path:   "users/1",
			Method: "GET",
			Headers: map[string]string{
				"X-Custom": "value",
			},
			Arguments: map[string]string{
				"query": "test",
			},
		}

		cfg := config.Setting{
			URL: server.URL,
			Headers: map[string]string{
				"Authorization": "Bearer test-token",
			},
		}

		httpReq, err := ashttpRequest.ToHTTPRequest(cfg)
		require.NoError(t, err, "ToHTTPRequest() should not fail")

		responseBody, err := Execute(httpReq)
		require.NoError(t, err, "Execute() should not fail")

		require.Equal(t, expectedResponse, string(responseBody), "Response body should match expected value")
	})
}

func BenchmarkToHTTPRequest(b *testing.B) {
	request := Request{
		Path:   "api/v1/users/123",
		Method: "GET",
		Headers: map[string]string{
			"X-Test": "value",
		},
		Arguments: map[string]string{
			"name":  "John",
			"email": "john@example.com",
		},
	}

	setting := config.Setting{
		URL: "https://api.example.com",
		Headers: map[string]string{
			"Authorization": "Bearer token",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := request.ToHTTPRequest(setting)
		require.NoError(b, err, "ToHTTPRequest() should not fail")
	}
}

func BenchmarkExecute(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"result": "success"}`)
	}))
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(b, err, "Should be able to create test request")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Execute(req)
		require.NoError(b, err, "Execute() should not fail")
	}
}

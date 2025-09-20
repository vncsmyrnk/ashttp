package http

import (
	"encoding/json"
	"fmt"
	"io"
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
		config          config.Config
		expectedURL     string
		expectedMethod  string
		expectedBody    string
		expectedHeaders map[string]string
		expectError     bool
	}{
		{
			name: "basic request with no headers or body",
			request: Request{
				Path:    "users",
				Headers: nil,
				Body:    nil,
			},
			config: config.Config{
				Domain:  "https://api.example.com",
				Headers: nil,
			},
			expectedURL:    "https://api.example.com/users",
			expectedMethod: "GET",
			expectedBody:   "null",
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
			},
			expectError: false,
		},
		{
			name: "request with path and config headers",
			request: Request{
				Path:    "posts/123",
				Headers: nil,
				Body:    nil,
			},
			config: config.Config{
				Domain: "https://jsonplaceholder.typicode.com",
				Headers: map[string]string{
					"Authorization": "Bearer token123",
					"User-Agent":    "ashttp/1.0",
				},
			},
			expectedURL:    "https://jsonplaceholder.typicode.com/posts/123",
			expectedMethod: "GET",
			expectedBody:   "null",
			expectedHeaders: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token123",
				"User-Agent":    "ashttp/1.0",
			},
			expectError: false,
		},
		{
			name: "request with custom headers",
			request: Request{
				Path: "api/v1/data",
				Headers: map[string]string{
					"X-Custom-Header": "custom-value",
					"Accept":          "application/json",
				},
				Body: nil,
			},
			config: config.Config{
				Domain: "https://localhost:8080",
				Headers: map[string]string{
					"Authorization": "Basic dXNlcjpwYXNz",
				},
			},
			expectedURL:    "https://localhost:8080/api/v1/data",
			expectedMethod: "GET",
			expectedBody:   "null",
			expectedHeaders: map[string]string{
				"Content-Type":    "application/json",
				"Authorization":   "Basic dXNlcjpwYXNz",
				"X-Custom-Header": "custom-value",
				"Accept":          "application/json",
			},
			expectError: false,
		},
		{
			name: "request with JSON body",
			request: Request{
				Path:    "users",
				Headers: nil,
				Body: map[string]any{
					"name":  "John Doe",
					"email": "john@example.com",
					"age":   30,
				},
			},
			config: config.Config{
				Domain:  "https://api.example.com",
				Headers: nil,
			},
			expectedURL:    "https://api.example.com/users",
			expectedMethod: "GET",
			expectedBody:   `{"age":30,"email":"john@example.com","name":"John Doe"}`,
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
			},
			expectError: false,
		},
		{
			name: "request with complex nested body",
			request: Request{
				Path:    "complex",
				Headers: nil,
				Body: map[string]any{
					"user": map[string]any{
						"name":    "Jane",
						"details": []string{"admin", "user"},
					},
					"meta": map[string]any{
						"version": 1.2,
						"active":  true,
					},
				},
			},
			config: config.Config{
				Domain:  "https://test.com",
				Headers: nil,
			},
			expectedURL:    "https://test.com/complex",
			expectedMethod: "GET",
			expectedBody:   `{"meta":{"active":true,"version":1.2},"user":{"details":["admin","user"],"name":"Jane"}}`,
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
			},
			expectError: false,
		},
		{
			name: "request with unmarshalable body should fail",
			request: Request{
				Path:    "test",
				Headers: nil,
				Body: map[string]any{
					"invalid": make(chan int), // channels cannot be marshaled to JSON
				},
			},
			config: config.Config{
				Domain:  "https://example.com",
				Headers: nil,
			},
			expectError: true,
		},
		{
			name: "empty path",
			request: Request{
				Path:    "",
				Headers: nil,
				Body:    nil,
			},
			config: config.Config{
				Domain:  "https://api.example.com",
				Headers: nil,
			},
			expectedURL:    "https://api.example.com/",
			expectedMethod: "GET",
			expectedBody:   "null",
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
			},
			expectError: false,
		},
		{
			name: "headers override - request headers take precedence",
			request: Request{
				Path: "override-test",
				Headers: map[string]string{
					"Authorization": "Bearer new-token", // Should override config auth
					"Custom-Header": "request-value",
				},
				Body: nil,
			},
			config: config.Config{
				Domain: "https://api.test.com",
				Headers: map[string]string{
					"Authorization":  "Bearer old-token", // Should be overridden
					"Default-Header": "config-value",
				},
			},
			expectedURL:    "https://api.test.com/override-test",
			expectedMethod: "GET",
			expectedBody:   "null",
			expectedHeaders: map[string]string{
				"Content-Type":   "application/json",
				"Authorization":  "Bearer new-token", // Request header wins
				"Custom-Header":  "request-value",
				"Default-Header": "config-value",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := tt.request.ToHTTPRequest(tt.config)

			if tt.expectError {
				require.Error(t, err, "ToHTTPRequest() should return an error")
				return
			}

			require.NoError(t, err, "ToHTTPRequest() should not return an error")
			require.Equal(t, tt.expectedURL, req.URL.String(), "URL should match expected value")
			require.Equal(t, tt.expectedMethod, req.Method, "HTTP method should match expected value")

			if req.Body != nil {
				bodyBytes, err := io.ReadAll(req.Body)
				require.NoError(t, err, "Should be able to read request body")

				var actualBody any
				var expectedBody any

				err = json.Unmarshal(bodyBytes, &actualBody)
				require.NoError(t, err, "Should be able to unmarshal actual body")

				err = json.Unmarshal([]byte(tt.expectedBody), &expectedBody)
				require.NoError(t, err, "Should be able to unmarshal expected body")

				require.Equal(t, expectedBody, actualBody, "Request body should match expected value")
			}

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
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, expectedResponse)
		}))
		defer server.Close()

		ashttpRequest := Request{
			Path: "users/1",
			Headers: map[string]string{
				"X-Custom": "value",
			},
			Body: map[string]any{
				"query": "test",
			},
		}

		cfg := config.Config{
			Domain: server.URL,
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
		Path: "api/v1/users/123",
		Headers: map[string]string{
			"X-Test": "value",
		},
		Body: map[string]any{
			"name":  "John",
			"email": "john@example.com",
		},
	}

	cfg := config.Config{
		Domain: "https://api.example.com",
		Headers: map[string]string{
			"Authorization": "Bearer token",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := request.ToHTTPRequest(cfg)
		require.NoError(b, err, "ToHTTPRequest() should not fail")
	}
}

func BenchmarkExecute(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

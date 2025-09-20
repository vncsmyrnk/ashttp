package http

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueryString_ToURL(t *testing.T) {
	tests := []struct {
		name        string
		queryString QueryString
		expectedURL string
	}{
		{
			name:        "nil query string",
			queryString: nil,
			expectedURL: "",
		},
		{
			name:        "empty query string",
			queryString: QueryString{},
			expectedURL: "",
		},
		{
			name: "single query parameter",
			queryString: QueryString{
				"key": "value",
			},
			expectedURL: "key=value",
		},
		{
			name: "multiple query parameters",
			queryString: QueryString{
				"name": "john",
				"age":  "30",
				"city": "newyork",
			},
			expectedURL: "age=30&city=newyork&name=john", // Note: map iteration order is not guaranteed, but we'll test for content
		},
		{
			name: "query parameters with special characters",
			queryString: QueryString{
				"search": "hello world",
				"filter": "type=user",
			},
			expectedURL: "filter=type=user&search=hello world",
		},
		{
			name: "query parameters with empty values",
			queryString: QueryString{
				"empty": "",
				"null":  "",
			},
			expectedURL: "empty=&null=",
		},
		{
			name: "single character values",
			queryString: QueryString{
				"a": "1",
				"b": "2",
			},
			expectedURL: "a=1&b=2",
		},
		{
			name: "numeric-like keys and values",
			queryString: QueryString{
				"123": "456",
				"789": "abc",
			},
			expectedURL: "123=456&789=abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.queryString.ToURL()

			if tt.expectedURL == "" {
				require.Equal(t, tt.expectedURL, result, "Empty query string should return empty string")
				return
			}

			expectedPairs := splitQueryString(tt.expectedURL)
			actualPairs := splitQueryString(result)

			require.ElementsMatch(t, expectedPairs, actualPairs, "Query parameters should match regardless of order")
		})
	}
}

func TestPathComponents_ToURL(t *testing.T) {
	tests := []struct {
		name           string
		pathComponents PathComponents
		expectedURL    string
	}{
		{
			name:           "nil path components",
			pathComponents: nil,
			expectedURL:    "",
		},
		{
			name:           "empty path components",
			pathComponents: PathComponents{},
			expectedURL:    "",
		},
		{
			name:           "single path component",
			pathComponents: PathComponents{"users"},
			expectedURL:    "users",
		},
		{
			name:           "multiple path components",
			pathComponents: PathComponents{"api", "v1", "users"},
			expectedURL:    "api/v1/users",
		},
		{
			name:           "path components with numbers",
			pathComponents: PathComponents{"users", "123", "profile"},
			expectedURL:    "users/123/profile",
		},
		{
			name:           "path components with special characters",
			pathComponents: PathComponents{"search", "hello-world", "results"},
			expectedURL:    "search/hello-world/results",
		},
		{
			name:           "path components with empty strings",
			pathComponents: PathComponents{"api", "", "users"},
			expectedURL:    "api//users",
		},
		{
			name:           "single empty string component",
			pathComponents: PathComponents{""},
			expectedURL:    "",
		},
		{
			name:           "multiple empty string components",
			pathComponents: PathComponents{"", "", ""},
			expectedURL:    "//",
		},
		{
			name:           "mixed content path components",
			pathComponents: PathComponents{"api", "v2", "users", "john-doe", "posts", "recent"},
			expectedURL:    "api/v2/users/john-doe/posts/recent",
		},
		{
			name:           "path components with underscores and dashes",
			pathComponents: PathComponents{"user_management", "get-profile", "admin_panel"},
			expectedURL:    "user_management/get-profile/admin_panel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pathComponents.ToURL()
			require.Equal(t, tt.expectedURL, result, "Path components should be joined correctly")
		})
	}
}

func TestPath(t *testing.T) {
	tests := []struct {
		name           string
		pathComponents PathComponents
		queryString    QueryString
		expectedURL    string
	}{
		{
			name:           "nil path and nil query",
			pathComponents: nil,
			queryString:    nil,
			expectedURL:    "",
		},
		{
			name:           "empty path and empty query",
			pathComponents: PathComponents{},
			queryString:    QueryString{},
			expectedURL:    "",
		},
		{
			name:           "path only, no query",
			pathComponents: PathComponents{"api", "users"},
			queryString:    nil,
			expectedURL:    "api/users",
		},
		{
			name:           "path only, empty query",
			pathComponents: PathComponents{"posts", "123"},
			queryString:    QueryString{},
			expectedURL:    "posts/123",
		},
		{
			name:           "empty path with query",
			pathComponents: PathComponents{},
			queryString:    QueryString{"search": "test"},
			expectedURL:    "?search=test",
		},
		{
			name:           "nil path with query",
			pathComponents: nil,
			queryString:    QueryString{"filter": "active"},
			expectedURL:    "?filter=active",
		},
		{
			name:           "path and single query parameter",
			pathComponents: PathComponents{"api", "v1", "users"},
			queryString:    QueryString{"limit": "10"},
			expectedURL:    "api/v1/users?limit=10",
		},
		{
			name:           "path and multiple query parameters",
			pathComponents: PathComponents{"search"},
			queryString: QueryString{
				"q":     "golang",
				"page":  "1",
				"limit": "20",
			},
			expectedURL: "search?", // We'll verify the query part separately due to map ordering
		},
		{
			name:           "complex path with complex query",
			pathComponents: PathComponents{"api", "v2", "users", "123", "posts"},
			queryString: QueryString{
				"include": "comments",
				"sort":    "date",
				"order":   "desc",
			},
			expectedURL: "api/v2/users/123/posts?",
		},
		{
			name:           "single path component with query",
			pathComponents: PathComponents{"dashboard"},
			queryString:    QueryString{"tab": "overview"},
			expectedURL:    "dashboard?tab=overview",
		},
		{
			name:           "path with empty string component and query",
			pathComponents: PathComponents{"api", "", "users"},
			queryString:    QueryString{"active": "true"},
			expectedURL:    "api//users?active=true",
		},
		{
			name:           "path and query with empty values",
			pathComponents: PathComponents{"test"},
			queryString:    QueryString{"empty": ""},
			expectedURL:    "test?empty=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Path(tt.pathComponents, tt.queryString)

			if tt.queryString == nil || len(tt.queryString) == 0 {
				require.Equal(t, tt.expectedURL, result, "URL should match exactly when no query parameters")
				return
			}

			if strings.Contains(tt.expectedURL, "?") && strings.HasSuffix(tt.expectedURL, "?") {
				expectedPathPart := strings.TrimSuffix(tt.expectedURL, "?")

				require.Contains(t, result, "?", "URL should contain query separator")
				parts := strings.Split(result, "?")
				require.Len(t, parts, 2, "URL should have exactly one query separator")

				actualPathPart := parts[0]
				actualQueryPart := parts[1]

				require.Equal(t, expectedPathPart, actualPathPart, "Path part should match expected")

				expectedPairs := convertQueryStringToPairs(tt.queryString)
				actualPairs := splitQueryString(actualQueryPart)
				require.ElementsMatch(t, expectedPairs, actualPairs, "Query parameters should match")
			} else {
				if strings.Contains(result, "?") && len(tt.queryString) > 1 {
					parts := strings.Split(result, "?")
					expectedPathPart := strings.Split(tt.expectedURL, "?")[0]
					require.Equal(t, expectedPathPart, parts[0], "Path part should match")

					expectedPairs := convertQueryStringToPairs(tt.queryString)
					actualPairs := splitQueryString(parts[1])
					require.ElementsMatch(t, expectedPairs, actualPairs, "Query parameters should match")
				} else {
					require.Equal(t, tt.expectedURL, result, "URL should match exactly")
				}
			}
		})
	}
}

func splitQueryString(queryString string) []string {
	if queryString == "" {
		return []string{}
	}
	return strings.Split(queryString, "&")
}

func convertQueryStringToPairs(qs QueryString) []string {
	pairs := make([]string, 0, len(qs))
	for k, v := range qs {
		pairs = append(pairs, k+"="+v)
	}
	return pairs
}

func BenchmarkQueryString_ToURL(b *testing.B) {
	queryString := QueryString{
		"search": "golang programming",
		"page":   "1",
		"limit":  "50",
		"sort":   "date",
		"order":  "desc",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = queryString.ToURL()
	}
}

func BenchmarkPathComponents_ToURL(b *testing.B) {
	pathComponents := PathComponents{"api", "v1", "users", "123", "posts", "recent"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pathComponents.ToURL()
	}
}

func BenchmarkPath(b *testing.B) {
	pathComponents := PathComponents{"api", "v2", "users", "profile"}
	queryString := QueryString{
		"include": "settings,preferences",
		"format":  "json",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Path(pathComponents, queryString)
	}
}

func TestQueryString_ToURL_EdgeCases(t *testing.T) {
	t.Run("large query string", func(t *testing.T) {
		largeQueryString := make(QueryString)
		for i := 0; i < 100; i++ {
			largeQueryString[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
		}

		result := largeQueryString.ToURL()
		require.NotEmpty(t, result, "Large query string should not be empty")
		require.Contains(t, result, "key0=value0", "Should contain first key-value pair")
		require.Contains(t, result, "key99=value99", "Should contain last key-value pair")

		separatorCount := strings.Count(result, "&")
		require.Equal(t, 99, separatorCount, "Should have correct number of separators")
	})
}

func TestPathComponents_ToURL_EdgeCases(t *testing.T) {
	t.Run("large path components", func(t *testing.T) {
		largePathComponents := make(PathComponents, 50)
		for i := 0; i < 50; i++ {
			largePathComponents[i] = fmt.Sprintf("segment%d", i)
		}

		result := largePathComponents.ToURL()
		require.NotEmpty(t, result, "Large path should not be empty")
		require.True(t, strings.HasPrefix(result, "segment0"), "Should start with first segment")
		require.True(t, strings.HasSuffix(result, "segment49"), "Should end with last segment")

		separatorCount := strings.Count(result, "/")
		require.Equal(t, 49, separatorCount, "Should have correct number of separators")
	})
}

func TestPath_EdgeCases(t *testing.T) {
	t.Run("very long combined URL", func(t *testing.T) {
		longPath := make(PathComponents, 20)
		for i := 0; i < 20; i++ {
			longPath[i] = fmt.Sprintf("very-long-path-segment-number-%d", i)
		}

		longQuery := make(QueryString)
		for i := 0; i < 20; i++ {
			longQuery[fmt.Sprintf("very-long-query-parameter-key-%d", i)] = fmt.Sprintf("very-long-query-parameter-value-%d", i)
		}

		result := Path(longPath, longQuery)
		require.NotEmpty(t, result, "Long combined URL should not be empty")
		require.Contains(t, result, "?", "Should contain query separator")

		parts := strings.Split(result, "?")
		require.Len(t, parts, 2, "Should have path and query parts")
		require.NotEmpty(t, parts[0], "Path part should not be empty")
		require.NotEmpty(t, parts[1], "Query part should not be empty")
	})
}

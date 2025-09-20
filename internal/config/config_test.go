package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetConfigs(t *testing.T) {
	tests := []struct {
		name           string
		setupMockFile  func(t *testing.T) (cleanup func())
		expectedResult ConfigByDomainAlias
		expectError    bool
	}{
		{
			name: "successful get configs with default config",
			setupMockFile: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				mockPath := filepath.Join(tmpDir, "config.json")

				originalPath := defaultFilePath
				defaultFilePath = mockPath

				return func() {
					defaultFilePath = originalPath
				}
			},
			expectedResult: ConfigByDomainAlias{
				DomainAlias("httpbin"): Config{
					Domain: "https://httpbin.dev/anything",
					Headers: map[string]string{
						"authorization": "123",
					},
				},
			},
			expectError: false,
		},
		{
			name: "successful get configs with custom config",
			setupMockFile: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				mockPath := filepath.Join(tmpDir, "config.json")

				customConfig := `{
					"api": {
						"url": "https://api.example.com",
						"defaultHeaders": {
							"Authorization": "Bearer token123",
							"Content-Type": "application/json"
						}
					},
					"staging": {
						"url": "https://staging.example.com",
						"defaultHeaders": {
							"X-Environment": "staging"
						}
					}
				}`

				err := os.WriteFile(mockPath, []byte(customConfig), 0644)
				require.NoError(t, err)

				originalPath := defaultFilePath
				defaultFilePath = mockPath

				return func() {
					defaultFilePath = originalPath
				}
			},
			expectedResult: ConfigByDomainAlias{
				DomainAlias("api"): Config{
					Domain: "https://api.example.com",
					Headers: map[string]string{
						"Authorization": "Bearer token123",
						"Content-Type":  "application/json",
					},
				},
				DomainAlias("staging"): Config{
					Domain: "https://staging.example.com",
					Headers: map[string]string{
						"X-Environment": "staging",
					},
				},
			},
			expectError: false,
		},
		{
			name: "empty config file",
			setupMockFile: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				mockPath := filepath.Join(tmpDir, "config.json")

				err := os.WriteFile(mockPath, []byte("{}"), 0644)
				require.NoError(t, err)

				originalPath := defaultFilePath
				defaultFilePath = mockPath

				return func() {
					defaultFilePath = originalPath
				}
			},
			expectedResult: ConfigByDomainAlias{},
			expectError:    false,
		},
		{
			name: "invalid config file causes error",
			setupMockFile: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				mockPath := filepath.Join(tmpDir, "config.json")

				err := os.WriteFile(mockPath, []byte("invalid json"), 0644)
				require.NoError(t, err)

				originalPath := defaultFilePath
				defaultFilePath = mockPath

				return func() {
					defaultFilePath = originalPath
				}
			},
			expectedResult: ConfigByDomainAlias{},
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setupMockFile(t)
			defer cleanup()

			result, err := GetConfigs()

			if tt.expectError {
				require.Error(t, err)
				require.Equal(t, ConfigByDomainAlias{}, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestConfigsFromExternalConfigs(t *testing.T) {
	tests := []struct {
		name            string
		externalConfigs ExternalConfig
		expectedResult  ConfigByDomainAlias
	}{
		{
			name: "single domain conversion",
			externalConfigs: ExternalConfig{
				"api": ExternalConfigDomainAlias{
					URL: "https://api.example.com",
					DefaultHeaders: map[string]string{
						"Authorization": "Bearer token123",
					},
				},
			},
			expectedResult: ConfigByDomainAlias{
				DomainAlias("api"): Config{
					Domain: "https://api.example.com",
					Headers: map[string]string{
						"Authorization": "Bearer token123",
					},
				},
			},
		},
		{
			name: "multiple domains conversion",
			externalConfigs: ExternalConfig{
				"api": ExternalConfigDomainAlias{
					URL: "https://api.example.com",
					DefaultHeaders: map[string]string{
						"Authorization": "Bearer token123",
						"Content-Type":  "application/json",
					},
				},
				"staging": ExternalConfigDomainAlias{
					URL: "https://staging.example.com",
					DefaultHeaders: map[string]string{
						"X-Environment": "staging",
					},
				},
				"localhost": ExternalConfigDomainAlias{
					URL: "http://localhost:8080",
					DefaultHeaders: map[string]string{
						"X-Debug": "true",
					},
				},
			},
			expectedResult: ConfigByDomainAlias{
				DomainAlias("api"): Config{
					Domain: "https://api.example.com",
					Headers: map[string]string{
						"Authorization": "Bearer token123",
						"Content-Type":  "application/json",
					},
				},
				DomainAlias("staging"): Config{
					Domain: "https://staging.example.com",
					Headers: map[string]string{
						"X-Environment": "staging",
					},
				},
				DomainAlias("localhost"): Config{
					Domain: "http://localhost:8080",
					Headers: map[string]string{
						"X-Debug": "true",
					},
				},
			},
		},
		{
			name:            "empty external config",
			externalConfigs: ExternalConfig{},
			expectedResult:  ConfigByDomainAlias{},
		},
		{
			name: "domain with empty headers",
			externalConfigs: ExternalConfig{
				"simple": ExternalConfigDomainAlias{
					URL:            "https://simple.example.com",
					DefaultHeaders: map[string]string{},
				},
			},
			expectedResult: ConfigByDomainAlias{
				DomainAlias("simple"): Config{
					Domain:  "https://simple.example.com",
					Headers: map[string]string{},
				},
			},
		},
		{
			name: "domain with nil headers",
			externalConfigs: ExternalConfig{
				"minimal": ExternalConfigDomainAlias{
					URL:            "https://minimal.example.com",
					DefaultHeaders: nil,
				},
			},
			expectedResult: ConfigByDomainAlias{
				DomainAlias("minimal"): Config{
					Domain:  "https://minimal.example.com",
					Headers: nil,
				},
			},
		},
		{
			name: "domains with special characters in alias",
			externalConfigs: ExternalConfig{
				"api-v1": ExternalConfigDomainAlias{
					URL: "https://api-v1.example.com",
					DefaultHeaders: map[string]string{
						"Version": "v1",
					},
				},
				"test_env": ExternalConfigDomainAlias{
					URL: "https://test-env.example.com",
					DefaultHeaders: map[string]string{
						"Environment": "test",
					},
				},
			},
			expectedResult: ConfigByDomainAlias{
				DomainAlias("api-v1"): Config{
					Domain: "https://api-v1.example.com",
					Headers: map[string]string{
						"Version": "v1",
					},
				},
				DomainAlias("test_env"): Config{
					Domain: "https://test-env.example.com",
					Headers: map[string]string{
						"Environment": "test",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := configsFromExternalConfigs(tt.externalConfigs)
			require.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestConfigTypes(t *testing.T) {
	t.Run("DomainAlias type conversion", func(t *testing.T) {
		alias := DomainAlias("test")
		require.Equal(t, "test", string(alias))

		// Test that it can be used as map key
		configs := make(ConfigByDomainAlias)
		configs[alias] = Config{
			Domain:  "https://test.com",
			Headers: map[string]string{"key": "value"},
		}

		retrieved, exists := configs[alias]
		require.True(t, exists)
		require.Equal(t, "https://test.com", retrieved.Domain)
		require.Equal(t, map[string]string{"key": "value"}, retrieved.Headers)
	})

	t.Run("Config struct initialization", func(t *testing.T) {
		config := Config{
			Domain: "https://example.com",
			Headers: map[string]string{
				"Authorization": "Bearer token",
				"Content-Type":  "application/json",
			},
		}

		require.Equal(t, "https://example.com", config.Domain)
		require.Equal(t, "Bearer token", config.Headers["Authorization"])
		require.Equal(t, "application/json", config.Headers["Content-Type"])
	})

	t.Run("ConfigByDomainAlias map operations", func(t *testing.T) {
		configs := ConfigByDomainAlias{
			"test1": Config{Domain: "https://test1.com", Headers: map[string]string{}},
			"test2": Config{Domain: "https://test2.com", Headers: map[string]string{}},
		}

		require.Len(t, configs, 2)
		require.Contains(t, configs, DomainAlias("test1"))
		require.Contains(t, configs, DomainAlias("test2"))

		// Test iteration
		count := 0
		for alias, config := range configs {
			require.NotEmpty(t, string(alias))
			require.NotEmpty(t, config.Domain)
			count++
		}
		require.Equal(t, 2, count)
	})
}

func TestGetConfigsIntegration(t *testing.T) {
	t.Run("end-to-end config loading and transformation", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockPath := filepath.Join(tmpDir, "config.json")

		// Create a comprehensive config file
		configContent := `{
			"production": {
				"url": "https://api.prod.example.com",
				"defaultHeaders": {
					"Authorization": "Bearer prod-token",
					"Content-Type": "application/json",
					"X-API-Version": "v2"
				}
			},
			"development": {
				"url": "http://localhost:3000",
				"defaultHeaders": {
					"X-Debug": "true"
				}
			}
		}`

		err := os.WriteFile(mockPath, []byte(configContent), 0644)
		require.NoError(t, err)

		// Override the defaultFilePath for this test
		originalPath := defaultFilePath
		defaultFilePath = mockPath
		defer func() {
			defaultFilePath = originalPath
		}()

		// Test GetConfigs
		configs, err := GetConfigs()
		require.NoError(t, err)
		require.Len(t, configs, 2)

		// Verify production config
		prodConfig, exists := configs[DomainAlias("production")]
		require.True(t, exists)
		require.Equal(t, "https://api.prod.example.com", prodConfig.Domain)
		require.Equal(t, map[string]string{
			"Authorization": "Bearer prod-token",
			"Content-Type":  "application/json",
			"X-API-Version": "v2",
		}, prodConfig.Headers)

		// Verify development config
		devConfig, exists := configs[DomainAlias("development")]
		require.True(t, exists)
		require.Equal(t, "http://localhost:3000", devConfig.Domain)
		require.Equal(t, map[string]string{
			"X-Debug": "true",
		}, devConfig.Headers)
	})
}

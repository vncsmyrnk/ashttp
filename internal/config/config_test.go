package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSettings(t *testing.T) {
	tests := []struct {
		name           string
		setupMockFile  func(t *testing.T) (cleanup func())
		expectedResult SettingByDomainAlias
		expectError    bool
	}{
		{
			name: "successful get settings with default config",
			setupMockFile: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				mockPath := filepath.Join(tmpDir, "config.json")

				originalPath := defaultFilePath
				defaultFilePath = mockPath

				return func() {
					defaultFilePath = originalPath
				}
			},
			expectedResult: SettingByDomainAlias{
				DomainAlias("httpbin"): Setting{
					Domain: "https://httpbin.dev/anything",
					Headers: map[string]string{
						"authorization": "123",
					},
				},
			},
			expectError: false,
		},
		{
			name: "successful get settings with custom config",
			setupMockFile: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				mockPath := filepath.Join(tmpDir, "config.json")

				customSetting := `{
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

				err := os.WriteFile(mockPath, []byte(customSetting), 0644)
				require.NoError(t, err)

				originalPath := defaultFilePath
				defaultFilePath = mockPath

				return func() {
					defaultFilePath = originalPath
				}
			},
			expectedResult: SettingByDomainAlias{
				DomainAlias("api"): Setting{
					Domain: "https://api.example.com",
					Headers: map[string]string{
						"Authorization": "Bearer token123",
						"Content-Type":  "application/json",
					},
				},
				DomainAlias("staging"): Setting{
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
			expectedResult: SettingByDomainAlias{},
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
			expectedResult: SettingByDomainAlias{},
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setupMockFile(t)
			defer cleanup()

			result, err := GetSettings()

			if tt.expectError {
				require.Error(t, err)
				require.Equal(t, SettingByDomainAlias{}, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestSettingsFromExternalSettings(t *testing.T) {
	tests := []struct {
		name             string
		externalSettings ExternalSetting
		expectedResult   SettingByDomainAlias
	}{
		{
			name: "single domain conversion",
			externalSettings: ExternalSetting{
				"api": ExternalSettingDomainAlias{
					URL: "https://api.example.com",
					DefaultHeaders: map[string]string{
						"Authorization": "Bearer token123",
					},
				},
			},
			expectedResult: SettingByDomainAlias{
				DomainAlias("api"): Setting{
					Domain: "https://api.example.com",
					Headers: map[string]string{
						"Authorization": "Bearer token123",
					},
				},
			},
		},
		{
			name: "multiple domains conversion",
			externalSettings: ExternalSetting{
				"api": ExternalSettingDomainAlias{
					URL: "https://api.example.com",
					DefaultHeaders: map[string]string{
						"Authorization": "Bearer token123",
						"Content-Type":  "application/json",
					},
				},
				"staging": ExternalSettingDomainAlias{
					URL: "https://staging.example.com",
					DefaultHeaders: map[string]string{
						"X-Environment": "staging",
					},
				},
				"localhost": ExternalSettingDomainAlias{
					URL: "http://localhost:8080",
					DefaultHeaders: map[string]string{
						"X-Debug": "true",
					},
				},
			},
			expectedResult: SettingByDomainAlias{
				DomainAlias("api"): Setting{
					Domain: "https://api.example.com",
					Headers: map[string]string{
						"Authorization": "Bearer token123",
						"Content-Type":  "application/json",
					},
				},
				DomainAlias("staging"): Setting{
					Domain: "https://staging.example.com",
					Headers: map[string]string{
						"X-Environment": "staging",
					},
				},
				DomainAlias("localhost"): Setting{
					Domain: "http://localhost:8080",
					Headers: map[string]string{
						"X-Debug": "true",
					},
				},
			},
		},
		{
			name:             "empty external config",
			externalSettings: ExternalSetting{},
			expectedResult:   SettingByDomainAlias{},
		},
		{
			name: "domain with empty headers",
			externalSettings: ExternalSetting{
				"simple": ExternalSettingDomainAlias{
					URL:            "https://simple.example.com",
					DefaultHeaders: map[string]string{},
				},
			},
			expectedResult: SettingByDomainAlias{
				DomainAlias("simple"): Setting{
					Domain:  "https://simple.example.com",
					Headers: map[string]string{},
				},
			},
		},
		{
			name: "domain with nil headers",
			externalSettings: ExternalSetting{
				"minimal": ExternalSettingDomainAlias{
					URL:            "https://minimal.example.com",
					DefaultHeaders: nil,
				},
			},
			expectedResult: SettingByDomainAlias{
				DomainAlias("minimal"): Setting{
					Domain:  "https://minimal.example.com",
					Headers: nil,
				},
			},
		},
		{
			name: "domains with special characters in alias",
			externalSettings: ExternalSetting{
				"api-v1": ExternalSettingDomainAlias{
					URL: "https://api-v1.example.com",
					DefaultHeaders: map[string]string{
						"Version": "v1",
					},
				},
				"test_env": ExternalSettingDomainAlias{
					URL: "https://test-env.example.com",
					DefaultHeaders: map[string]string{
						"Environment": "test",
					},
				},
			},
			expectedResult: SettingByDomainAlias{
				DomainAlias("api-v1"): Setting{
					Domain: "https://api-v1.example.com",
					Headers: map[string]string{
						"Version": "v1",
					},
				},
				DomainAlias("test_env"): Setting{
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
			result := settingsFromExternalSettings(tt.externalSettings)
			require.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestSettingTypes(t *testing.T) {
	t.Run("DomainAlias type conversion", func(t *testing.T) {
		alias := DomainAlias("test")
		require.Equal(t, "test", string(alias))

		settings := make(SettingByDomainAlias)
		settings[alias] = Setting{
			Domain:  "https://test.com",
			Headers: map[string]string{"key": "value"},
		}

		retrieved, exists := settings[alias]
		require.True(t, exists)
		require.Equal(t, "https://test.com", retrieved.Domain)
		require.Equal(t, map[string]string{"key": "value"}, retrieved.Headers)
	})

	t.Run("Setting struct initialization", func(t *testing.T) {
		setting := Setting{
			Domain: "https://example.com",
			Headers: map[string]string{
				"Authorization": "Bearer token",
				"Content-Type":  "application/json",
			},
		}

		require.Equal(t, "https://example.com", setting.Domain)
		require.Equal(t, "Bearer token", setting.Headers["Authorization"])
		require.Equal(t, "application/json", setting.Headers["Content-Type"])
	})

	t.Run("SettingByDomainAlias map operations", func(t *testing.T) {
		settings := SettingByDomainAlias{
			"test1": Setting{Domain: "https://test1.com", Headers: map[string]string{}},
			"test2": Setting{Domain: "https://test2.com", Headers: map[string]string{}},
		}

		require.Len(t, settings, 2)
		require.Contains(t, settings, DomainAlias("test1"))
		require.Contains(t, settings, DomainAlias("test2"))

		// Test iteration
		count := 0
		for alias, setting := range settings {
			require.NotEmpty(t, string(alias))
			require.NotEmpty(t, setting.Domain)
			count++
		}
		require.Equal(t, 2, count)
	})
}

func TestGetSettingsIntegration(t *testing.T) {
	t.Run("end-to-end config loading and transformation", func(t *testing.T) {
		tmpDir := t.TempDir()
		mockPath := filepath.Join(tmpDir, "config.json")

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

		originalPath := defaultFilePath
		defaultFilePath = mockPath
		defer func() {
			defaultFilePath = originalPath
		}()

		settings, err := GetSettings()
		require.NoError(t, err)
		require.Len(t, settings, 2)

		prodSetting, exists := settings[DomainAlias("production")]
		require.True(t, exists)
		require.Equal(t, "https://api.prod.example.com", prodSetting.Domain)
		require.Equal(t, map[string]string{
			"Authorization": "Bearer prod-token",
			"Content-Type":  "application/json",
			"X-API-Version": "v2",
		}, prodSetting.Headers)

		devSetting, exists := settings[DomainAlias("development")]
		require.True(t, exists)
		require.Equal(t, "http://localhost:3000", devSetting.Domain)
		require.Equal(t, map[string]string{
			"X-Debug": "true",
		}, devSetting.Headers)
	})
}

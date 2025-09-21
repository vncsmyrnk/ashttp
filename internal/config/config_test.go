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
		expectedResult SettingByURLAlias
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
			expectedResult: SettingByURLAlias{
				URLAlias("httpbin"): Setting{
					URL: "https://httpbin.dev/anything",
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
			expectedResult: SettingByURLAlias{
				URLAlias("api"): Setting{
					URL: "https://api.example.com",
					Headers: map[string]string{
						"Authorization": "Bearer token123",
						"Content-Type":  "application/json",
					},
				},
				URLAlias("staging"): Setting{
					URL: "https://staging.example.com",
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
			expectedResult: SettingByURLAlias{},
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
			expectedResult: SettingByURLAlias{},
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
				require.Equal(t, SettingByURLAlias{}, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestGetDefaultConfigPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err, "should be able to get user home directory")

	expectedPath := filepath.Join(homeDir, ".config", "ashttp", "config.json")
	actualPath := GetDefaultConfigPath()

	require.Equal(t, expectedPath, actualPath, "default config path should match expected path")
}

func TestSettingsFromExternalSettings(t *testing.T) {
	tests := []struct {
		name             string
		externalSettings ExternalSetting
		expectedResult   SettingByURLAlias
	}{
		{
			name: "single URL conversion",
			externalSettings: ExternalSetting{
				"api": ExternalSettingURLAlias{
					URL: "https://api.example.com",
					DefaultHeaders: map[string]string{
						"Authorization": "Bearer token123",
					},
				},
			},
			expectedResult: SettingByURLAlias{
				URLAlias("api"): Setting{
					URL: "https://api.example.com",
					Headers: map[string]string{
						"Authorization": "Bearer token123",
					},
				},
			},
		},
		{
			name: "multiple URLs conversion",
			externalSettings: ExternalSetting{
				"api": ExternalSettingURLAlias{
					URL: "https://api.example.com",
					DefaultHeaders: map[string]string{
						"Authorization": "Bearer token123",
						"Content-Type":  "application/json",
					},
				},
				"staging": ExternalSettingURLAlias{
					URL: "https://staging.example.com",
					DefaultHeaders: map[string]string{
						"X-Environment": "staging",
					},
				},
				"localhost": ExternalSettingURLAlias{
					URL: "http://localhost:8080",
					DefaultHeaders: map[string]string{
						"X-Debug": "true",
					},
				},
			},
			expectedResult: SettingByURLAlias{
				URLAlias("api"): Setting{
					URL: "https://api.example.com",
					Headers: map[string]string{
						"Authorization": "Bearer token123",
						"Content-Type":  "application/json",
					},
				},
				URLAlias("staging"): Setting{
					URL: "https://staging.example.com",
					Headers: map[string]string{
						"X-Environment": "staging",
					},
				},
				URLAlias("localhost"): Setting{
					URL: "http://localhost:8080",
					Headers: map[string]string{
						"X-Debug": "true",
					},
				},
			},
		},
		{
			name:             "empty external config",
			externalSettings: ExternalSetting{},
			expectedResult:   SettingByURLAlias{},
		},
		{
			name: "URL with empty headers",
			externalSettings: ExternalSetting{
				"simple": ExternalSettingURLAlias{
					URL:            "https://simple.example.com",
					DefaultHeaders: map[string]string{},
				},
			},
			expectedResult: SettingByURLAlias{
				URLAlias("simple"): Setting{
					URL:     "https://simple.example.com",
					Headers: map[string]string{},
				},
			},
		},
		{
			name: "URL with nil headers",
			externalSettings: ExternalSetting{
				"minimal": ExternalSettingURLAlias{
					URL:            "https://minimal.example.com",
					DefaultHeaders: nil,
				},
			},
			expectedResult: SettingByURLAlias{
				URLAlias("minimal"): Setting{
					URL:     "https://minimal.example.com",
					Headers: nil,
				},
			},
		},
		{
			name: "URLs with special characters in alias",
			externalSettings: ExternalSetting{
				"api-v1": ExternalSettingURLAlias{
					URL: "https://api-v1.example.com",
					DefaultHeaders: map[string]string{
						"Version": "v1",
					},
				},
				"test_env": ExternalSettingURLAlias{
					URL: "https://test-env.example.com",
					DefaultHeaders: map[string]string{
						"Environment": "test",
					},
				},
			},
			expectedResult: SettingByURLAlias{
				URLAlias("api-v1"): Setting{
					URL: "https://api-v1.example.com",
					Headers: map[string]string{
						"Version": "v1",
					},
				},
				URLAlias("test_env"): Setting{
					URL: "https://test-env.example.com",
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
	t.Run("URLAlias type conversion", func(t *testing.T) {
		alias := URLAlias("test")
		require.Equal(t, "test", string(alias))

		settings := make(SettingByURLAlias)
		settings[alias] = Setting{
			URL:     "https://test.com",
			Headers: map[string]string{"key": "value"},
		}

		retrieved, exists := settings[alias]
		require.True(t, exists)
		require.Equal(t, "https://test.com", retrieved.URL)
		require.Equal(t, map[string]string{"key": "value"}, retrieved.Headers)
	})

	t.Run("Setting struct initialization", func(t *testing.T) {
		setting := Setting{
			URL: "https://example.com",
			Headers: map[string]string{
				"Authorization": "Bearer token",
				"Content-Type":  "application/json",
			},
		}

		require.Equal(t, "https://example.com", setting.URL)
		require.Equal(t, "Bearer token", setting.Headers["Authorization"])
		require.Equal(t, "application/json", setting.Headers["Content-Type"])
	})

	t.Run("SettingByURLAlias map operations", func(t *testing.T) {
		settings := SettingByURLAlias{
			"test1": Setting{URL: "https://test1.com", Headers: map[string]string{}},
			"test2": Setting{URL: "https://test2.com", Headers: map[string]string{}},
		}

		require.Len(t, settings, 2)
		require.Contains(t, settings, URLAlias("test1"))
		require.Contains(t, settings, URLAlias("test2"))

		count := 0
		for alias, setting := range settings {
			require.NotEmpty(t, string(alias))
			require.NotEmpty(t, setting.URL)
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

		prodSetting, exists := settings[URLAlias("production")]
		require.True(t, exists)
		require.Equal(t, "https://api.prod.example.com", prodSetting.URL)
		require.Equal(t, map[string]string{
			"Authorization": "Bearer prod-token",
			"Content-Type":  "application/json",
			"X-API-Version": "v2",
		}, prodSetting.Headers)

		devSetting, exists := settings[URLAlias("development")]
		require.True(t, exists)
		require.Equal(t, "http://localhost:3000", devSetting.URL)
		require.Equal(t, map[string]string{
			"X-Debug": "true",
		}, devSetting.Headers)
	})
}

package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigFromFile(t *testing.T) {
	tests := []struct {
		name           string
		setupFile      func(t *testing.T) string
		expectedConfig ExternalConfig
		expectError    bool
	}{
		{
			name: "successful load existing config file",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.json")

				testConfig := ExternalConfig{
					"example": ExternalConfigDomainAlias{
						URL: "https://example.com",
						DefaultHeaders: map[string]string{
							"Content-Type": "application/json",
						},
					},
				}

				data, err := json.MarshalIndent(testConfig, "", "  ")
				require.NoError(t, err)

				err = os.WriteFile(configPath, data, 0644)
				require.NoError(t, err)

				return configPath
			},
			expectedConfig: ExternalConfig{
				"example": ExternalConfigDomainAlias{
					URL: "https://example.com",
					DefaultHeaders: map[string]string{
						"Content-Type": "application/json",
					},
				},
			},
			expectError: false,
		},
		{
			name: "file does not exist - creates default config",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "nonexistent.json")
			},
			expectedConfig: defaultConfig,
			expectError:    false,
		},
		{
			name: "invalid json file",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.json")

				err := os.WriteFile(configPath, []byte("invalid json content"), 0644)
				require.NoError(t, err)

				return configPath
			},
			expectedConfig: nil,
			expectError:    true,
		},
		{
			name: "empty json file",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.json")

				err := os.WriteFile(configPath, []byte("{}"), 0644)
				require.NoError(t, err)

				return configPath
			},
			expectedConfig: ExternalConfig{},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := tt.setupFile(t)

			result, err := loadConfigFromFile(configPath)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedConfig, result)
			}
		})
	}
}

func TestCreateDefaultConfig(t *testing.T) {
	tests := []struct {
		name        string
		setupPath   func(t *testing.T) string
		expectError bool
	}{
		{
			name: "successful creation in new directory",
			setupPath: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "subdir", "config.json")
			},
			expectError: false,
		},
		{
			name: "successful creation in existing directory",
			setupPath: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "config.json")
			},
			expectError: false,
		},
		{
			name: "overwrite existing file",
			setupPath: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.json")

				err := os.WriteFile(configPath, []byte("existing content"), 0644)
				require.NoError(t, err)

				return configPath
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := tt.setupPath(t)

			err := createDefaultConfig(configPath)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				require.FileExists(t, configPath)

				data, err := os.ReadFile(configPath)
				require.NoError(t, err)

				var config ExternalConfig
				err = json.Unmarshal(data, &config)
				require.NoError(t, err)
				require.Equal(t, defaultConfig, config)

				expectedData, err := json.MarshalIndent(defaultConfig, "", "  ")
				require.NoError(t, err)
				require.Equal(t, expectedData, data)
			}
		})
	}
}

func TestLoadConfigIntegration(t *testing.T) {
	tests := []struct {
		name                string
		configContent       ExternalConfig
		expectedLoadSuccess bool
	}{
		{
			name: "complex config with multiple domains",
			configContent: ExternalConfig{
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
			},
			expectedLoadSuccess: true,
		},
		{
			name: "config with empty headers",
			configContent: ExternalConfig{
				"simple": ExternalConfigDomainAlias{
					URL:            "https://simple.example.com",
					DefaultHeaders: map[string]string{},
				},
			},
			expectedLoadSuccess: true,
		},
		{
			name: "config with nil headers",
			configContent: ExternalConfig{
				"minimal": ExternalConfigDomainAlias{
					URL:            "https://minimal.example.com",
					DefaultHeaders: nil,
				},
			},
			expectedLoadSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.json")

			data, err := json.MarshalIndent(tt.configContent, "", "  ")
			require.NoError(t, err)

			err = os.WriteFile(configPath, data, 0644)
			require.NoError(t, err)

			result, err := loadConfigFromFile(configPath)

			if tt.expectedLoadSuccess {
				require.NoError(t, err)
				require.Equal(t, tt.configContent, result)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDefaultConfigValues(t *testing.T) {
	require.NotEmpty(t, defaultConfig)
	require.Contains(t, defaultConfig, "httpbin")

	httpbinConfig := defaultConfig["httpbin"]
	require.Equal(t, "https://httpbin.dev/anything", httpbinConfig.URL)
	require.Equal(t, map[string]string{"authorization": "123"}, httpbinConfig.DefaultHeaders)
}

func TestFilePathVariables(t *testing.T) {
	require.NotEmpty(t, defaultFileFolder)
	require.NotEmpty(t, defaultFilePath)
	require.Contains(t, defaultFilePath, "config.json")
	require.Contains(t, defaultFileFolder, ".config")
	require.Contains(t, defaultFileFolder, "ashttp")
}

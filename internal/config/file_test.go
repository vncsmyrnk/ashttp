package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadSettingFromFile(t *testing.T) {
	tests := []struct {
		name            string
		setupFile       func(t *testing.T) string
		expectedSetting ExternalSetting
		expectError     bool
	}{
		{
			name: "successful load existing config file",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.json")

				testSetting := ExternalSetting{
					"example": ExternalSettingURLAlias{
						URL: "https://example.com",
						DefaultHeaders: map[string]string{
							"Content-Type": "application/json",
						},
					},
				}

				data, err := json.MarshalIndent(testSetting, "", "  ")
				require.NoError(t, err)

				err = os.WriteFile(configPath, data, 0644)
				require.NoError(t, err)

				return configPath
			},
			expectedSetting: ExternalSetting{
				"example": ExternalSettingURLAlias{
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
			expectedSetting: defaultSetting,
			expectError:     false,
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
			expectedSetting: nil,
			expectError:     true,
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
			expectedSetting: ExternalSetting{},
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := tt.setupFile(t)

			result, err := loadSettingFromFile(configPath)

			if tt.expectError {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedSetting, result)
			}
		})
	}
}

func TestCreateDefaultSetting(t *testing.T) {
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

			err := createDefaultSetting(configPath)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				require.FileExists(t, configPath)

				data, err := os.ReadFile(configPath)
				require.NoError(t, err)

				var setting ExternalSetting
				err = json.Unmarshal(data, &setting)
				require.NoError(t, err)
				require.Equal(t, defaultSetting, setting)

				expectedData, err := json.MarshalIndent(defaultSetting, "", "  ")
				require.NoError(t, err)
				require.Equal(t, expectedData, data)
			}
		})
	}
}

func TestLoadSettingIntegration(t *testing.T) {
	tests := []struct {
		name                string
		setting             ExternalSetting
		expectedLoadSuccess bool
	}{
		{
			name: "complex setting with multiple URLs",
			setting: ExternalSetting{
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
			},
			expectedLoadSuccess: true,
		},
		{
			name: "setting with empty headers",
			setting: ExternalSetting{
				"simple": ExternalSettingURLAlias{
					URL:            "https://simple.example.com",
					DefaultHeaders: map[string]string{},
				},
			},
			expectedLoadSuccess: true,
		},
		{
			name: "setting with nil headers",
			setting: ExternalSetting{
				"minimal": ExternalSettingURLAlias{
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

			data, err := json.MarshalIndent(tt.setting, "", "  ")
			require.NoError(t, err)

			err = os.WriteFile(configPath, data, 0644)
			require.NoError(t, err)

			result, err := loadSettingFromFile(configPath)

			if tt.expectedLoadSuccess {
				require.NoError(t, err)
				require.Equal(t, tt.setting, result)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestDefaultSettingValues(t *testing.T) {
	require.NotEmpty(t, defaultSetting)
	require.Contains(t, defaultSetting, "httpbin")

	httpbinSetting := defaultSetting["httpbin"]
	require.Equal(t, "https://httpbin.dev/anything", httpbinSetting.URL)
	require.Equal(t, map[string]string{"authorization": "123"}, httpbinSetting.DefaultHeaders)
}

func TestFilePathVariables(t *testing.T) {
	require.NotEmpty(t, defaultFileFolder)
	require.NotEmpty(t, defaultFilePath)
	require.Contains(t, defaultFilePath, "config.json")
	require.Contains(t, defaultFileFolder, ".config")
	require.Contains(t, defaultFileFolder, "ashttp")
}

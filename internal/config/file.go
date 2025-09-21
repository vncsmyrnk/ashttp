package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

type ExternalSettingURLAlias struct {
	URL            string            `json:"url"`
	DefaultHeaders map[string]string `json:"defaultHeaders"`
}

type ExternalSetting map[string]ExternalSettingURLAlias

var defaultFileFolder = path.Join(os.ExpandEnv("$HOME"), ".config", "ashttp")
var defaultFilePath = path.Join(defaultFileFolder, "config.json")

var defaultSetting = ExternalSetting{
	"httpbin": ExternalSettingURLAlias{
		URL: "https://httpbin.dev/anything",
		DefaultHeaders: map[string]string{
			"authorization": "123",
		},
	},
}

func loadSettingFromFile(filePath string) (ExternalSetting, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if createErr := createDefaultSetting(filePath); createErr != nil {
			return nil, createErr
		}
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var configs ExternalSetting
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return configs, nil
}

func createDefaultSetting(filePath string) error {
	data, err := json.MarshalIndent(defaultSetting, "", "  ")
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

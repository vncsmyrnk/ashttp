package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
)

type ExternalConfigDomainAlias struct {
	URL            string            `json:"url"`
	DefaultHeaders map[string]string `json:"defaultHeaders"`
}

type ExternalConfig map[string]ExternalConfigDomainAlias

var defaultFileFolder = path.Join(os.ExpandEnv("$HOME"), ".config", "ashttp")
var defaultFilePath = path.Join(defaultFileFolder, "config.json")

var defaultConfig = ExternalConfig{
	"httpbin": ExternalConfigDomainAlias{
		URL: "https://httpbin.dev/anything",
		DefaultHeaders: map[string]string{
			"authorization": "123",
		},
	},
}

func loadConfigFromFile(filePath string) (ExternalConfig, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		createDefaultConfig(filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var configs ExternalConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return configs, nil
}

func createDefaultConfig(filePath string) error {
	data, err := json.MarshalIndent(defaultConfig, "", "  ")
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

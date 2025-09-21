package config

type Config struct {
	Domain  string
	Headers map[string]string
}

type DomainAlias string

// nolint:revive
type ConfigByDomainAlias map[DomainAlias]Config

func GetConfigs() (ConfigByDomainAlias, error) {
	configs, err := loadConfigFromFile(defaultFilePath)
	if err != nil {
		return ConfigByDomainAlias{}, err
	}

	return configsFromExternalConfigs(configs), nil
}

func GetDefaultConfigPath() string {
	return defaultFilePath
}

func configsFromExternalConfigs(externalConfigs ExternalConfig) ConfigByDomainAlias {
	configs := make(ConfigByDomainAlias)
	for k, v := range externalConfigs {
		domainAlias := DomainAlias(k)
		configs[domainAlias] = Config{
			Domain:  v.URL,
			Headers: v.DefaultHeaders,
		}
	}

	return configs
}

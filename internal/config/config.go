package config

type Setting struct {
	Domain  string
	Headers map[string]string
}

type DomainAlias string

type SettingByDomainAlias map[DomainAlias]Setting

func GetSettings() (SettingByDomainAlias, error) {
	settings, err := loadSettingFromFile(defaultFilePath)
	if err != nil {
		return SettingByDomainAlias{}, err
	}

	return settingsFromExternalSettings(settings), nil
}

func GetDefaultSettingPath() string {
	return defaultFilePath
}

func settingsFromExternalSettings(externalSettings ExternalSetting) SettingByDomainAlias {
	settings := make(SettingByDomainAlias)
	for k, v := range externalSettings {
		domainAlias := DomainAlias(k)
		settings[domainAlias] = Setting{
			Domain:  v.URL,
			Headers: v.DefaultHeaders,
		}
	}

	return settings
}

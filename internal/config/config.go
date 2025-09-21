package config

type Setting struct {
	URL     string
	Headers map[string]string
}

type URLAlias string

type SettingByURLAlias map[URLAlias]Setting

func GetSettings() (SettingByURLAlias, error) {
	settings, err := loadSettingFromFile(defaultFilePath)
	if err != nil {
		return SettingByURLAlias{}, err
	}

	return settingsFromExternalSettings(settings), nil
}

func GetDefaultConfigPath() string {
	return defaultFilePath
}

func settingsFromExternalSettings(externalSettings ExternalSetting) SettingByURLAlias {
	settings := make(SettingByURLAlias)
	for k, v := range externalSettings {
		urlAlias := URLAlias(k)
		settings[urlAlias] = Setting{
			URL:     v.URL,
			Headers: v.DefaultHeaders,
		}
	}

	return settings
}

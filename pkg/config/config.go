package config

import "fmt"

type Config interface {
	WriteConfigToFile() error
	Set(key string, value []byte)
	Generate() (string, error)
}

type RequiredFile struct {
	Filename     string
	PathKey      string
	IncludeInEnv bool
}

func requiredSecretDataExists(required map[string]RequiredFile, secretData map[string][]byte, filetype string) error {
	for key, _ := range required {
		if _, ok := secretData[key]; !ok {
			return fmt.Errorf("can not generate %s config, secret missing required key: %s", filetype, key)
		}
	}
	return nil
}

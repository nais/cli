package command

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func quoteShellArgument(value string) string {
	if value != "" && strings.Trim(value, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_.") == "" {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func postgresGetCommandLine(name, team, environment, configFile string) string {
	command := "nais alpha postgres get " + quoteShellArgument(name)
	validDefaults := true
	if configFile != "" {
		absolute, err := filepath.Abs(configFile)
		if err != nil {
			validDefaults = false
		} else {
			configFile = absolute
		}
		configDir, err := os.UserConfigDir()
		if err != nil || filepath.Clean(configFile) != filepath.Join(configDir, "nais", "config.yaml") {
			command += " --config " + quoteShellArgument(configFile)
		}
	}
	defaults := viper.New()
	defaults.SetEnvPrefix("NAIS")
	defaults.AutomaticEnv()
	if configFile != "" && validDefaults {
		defaults.SetConfigFile(configFile)
		if err := defaults.ReadInConfig(); err != nil && !errors.Is(err, os.ErrNotExist) {
			validDefaults = false
		}
	}
	if !validDefaults || defaults.GetString("team") != team {
		command += " -t " + quoteShellArgument(team)
	}
	if !validDefaults || defaults.GetString("environment") != environment {
		command += " -e " + quoteShellArgument(environment)
	}
	return command
}

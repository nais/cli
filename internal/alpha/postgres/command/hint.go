package command

import (
	"os"
	"path/filepath"
	"strings"
)

func quoteShellArgument(value string) string {
	if value != "" && strings.Trim(value, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_.") == "" {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func postgresGetCommandLine(name, team, environment, configFile string) string {
	return postgresCommandLine("nais alpha postgres get "+quoteShellArgument(name), team, environment, configFile)
}

func branchListCommandLine(name, team, environment, configFile string) string {
	return postgresCommandLine("nais alpha postgres branch list "+quoteShellArgument(name), team, environment, configFile)
}

func postgresCommandLine(command, team, environment, configFile string) string {
	if configFile != "" {
		if absolute, err := filepath.Abs(configFile); err == nil {
			configFile = absolute
		}
		configDir, err := os.UserConfigDir()
		if err != nil || filepath.Clean(configFile) != filepath.Join(configDir, "nais", "config.yaml") {
			command += " --config " + quoteShellArgument(configFile)
		}
	}
	return command + " -t " + quoteShellArgument(team) + " -e " + quoteShellArgument(environment)
}

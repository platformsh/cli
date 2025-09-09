// Package alt manages instances of alternative CLI configurations.
package alt

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

type Alt struct {
	executablePath string
	comment        string
	target         string

	configPath string
	configNode *yaml.Node
}

func New(execPath, comment, target, configPath string, configNode *yaml.Node) *Alt {
	return &Alt{
		executablePath: execPath,
		comment:        comment,
		target:         target,
		configPath:     configPath,
		configNode:     configNode,
	}
}

// GenerateAndSave creates and saves the executable and config files needed for the alt CLI instance.
func (a *Alt) GenerateAndSave() error {
	configContent, err := yaml.Marshal(a.configNode)
	if err != nil {
		return err
	}
	if err := writeFile(a.configPath, configContent, 0o755, 0o644); err != nil {
		return err
	}
	executableContent := a.generateExecutable()
	return writeFile(a.executablePath, []byte(executableContent), 0o755, 0o755)
}

func (a *Alt) generateExecutable() string {
	if runtime.GOOS == "windows" {
		return ":: " + a.comment + "\r\n" +
			"@echo off\r\n" +
			"setlocal\r\n" +
			`set CLI_CONFIG_FILE=` + formatConfigPathForShell(a.configPath) + "\r\n" +
			a.target + " %*\r\n" +
			"endlocal\r\n"
	}

	return "#!/bin/sh\n" +
		"# " + a.comment + "\n" +
		"export CLI_CONFIG_FILE=" + formatConfigPathForShell(a.configPath) + "\n" +
		a.target + ` "$@"` + "\n"
}

func formatConfigPathForShell(configPath string) string {
	if runtime.GOOS == "windows" {
		if strings.HasPrefix(configPath, os.Getenv("AppData")) {
			return "%AppData%" + strings.TrimPrefix(configPath, os.Getenv("AppData"))
		}
		return configPath
	}
	vars := []string{"XDG_CONFIG_HOME", "HOME"}
	for _, v := range vars {
		val := os.Getenv(v)
		if val != "" && strings.HasPrefix(configPath, val) {
			return fmt.Sprintf(`"${%s}%s"`, v, strings.TrimPrefix(configPath, val))
		}
	}
	return `"` + configPath + `"`
}

func GetExecutableFileExtension() string {
	if runtime.GOOS == "windows" {
		return ".bat"
	}
	return ""
}

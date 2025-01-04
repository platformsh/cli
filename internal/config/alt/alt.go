// Package alt manages instances of alternative CLI configurations.
package alt

import (
	"io/fs"
	"os"
	"path/filepath"
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
	if err := writeFileAndDir(a.configPath, configContent, 0o755, 0o644); err != nil {
		return err
	}
	executableContent := a.generateExecutable()
	return writeFileAndDir(a.executablePath, []byte(executableContent), 0o755, 0o755)
}

func (a *Alt) generateExecutable() string {
	if runtime.GOOS == "windows" {
		return withLineEnding("\r\n", "@echo off",
			":: "+a.comment,
			`set "CLI_CONFIG_FILE=`+a.configPath+`"`,
			// TODO check if this works on Windows
			`for /f "tokens=1,* delims= " %%a in ("%*") do set ARGS_BUT_FIRST=%%b`,
			a.target+" %ARGS_BUT_FIRST%",
		)
	}

	return withLineEnding("\n", "#!/bin/sh",
		"# "+a.comment,
		"export CLI_CONFIG_FILE="+a.configPath,
		`[ "$#" -gt 1 ] && shift # Skip first argument`,
		a.target+` "$@"`,
	)
}

func withLineEnding(e string, lines ...string) string {
	return strings.Join(lines, e) + e
}

// Write or overwrite a file, creating its containing directory if necessary.
func writeFileAndDir(path string, content []byte, dirMode, fileMode fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), dirMode); err != nil {
		return err
	}
	if err := os.WriteFile(path, content, fileMode); err != nil {
		return err
	}
	return nil
}

func GetExecutableFileExtension() string {
	if runtime.GOOS == "windows" {
		return ".bat"
	}
	return ""
}

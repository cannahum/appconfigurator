package appconfigurator

import (
	"fmt"
	"os"
	"path"
)

func DefaultConfigurationDirectoryPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", nil
	}
	for {
		if _, err := os.Stat(path.Join(wd, "go.mod")); err == nil {
			break
		}
		parentDir := path.Dir(wd)
		if parentDir == wd {
			return "", fmt.Errorf("could not find project root (searching for go.mod)")
		}
		wd = parentDir
	}
	return path.Join(wd, "configuration"), nil
}

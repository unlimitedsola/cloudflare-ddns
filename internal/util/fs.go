package util

import (
	"os"
	"path/filepath"
)

func GetExecutableAbsolutePath() (string, error) {
	exec, err := os.Executable()
	if err != nil {
		return "", err
	}
	absPath, err := filepath.Abs(exec)
	if err != nil {
		return "", err
	}
	return absPath, nil
}

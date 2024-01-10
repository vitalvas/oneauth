package tools

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetHomeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return home, nil
}

func InHomeDir(items ...string) (string, error) {
	dir, err := GetHomeDir()
	if err != nil {
		return "", err
	}

	path := append([]string{dir}, items...)

	return filepath.Join(path...), nil
}

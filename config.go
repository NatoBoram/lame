package main

import (
	"fmt"
	"os"
)

func lameConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config dir: %w", err)
	}

	path := dir + string(os.PathSeparator) + "lame"
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return path, fmt.Errorf("failed to create `lame` config dir: %w", err)
	}

	return path, err
}

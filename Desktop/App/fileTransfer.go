package main

import (
	"os"
	"path/filepath"
)

func isValidFilePath(path string) bool {
	if path == "" {
		return false
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return false
	}

	if info.IsDir() {
		return false
	}

	return true
}

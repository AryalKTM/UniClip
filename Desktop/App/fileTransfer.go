package main

import (
	"os"
	"path/filepath"
)

func isValidFilePath(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	if path == "" {
		return false
	}

	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return false
	}

	if info.IsDir() {
		return false
	}
	return true
}

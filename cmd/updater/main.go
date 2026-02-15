package main

import (
	"log"
	"os"
	"path/filepath"
)

var version = "dev"

func main() {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("updater: failed to get executable path: %v", err)
	}

	execDir := filepath.Dir(execPath)

	if isWindowsService() {
		runAsService(execDir)
	} else {
		runWrapper(execDir)
	}
}

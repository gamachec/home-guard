//go:build !windows

package main

import "log"

func isWindowsService() bool { return false }

func runAsService(execDir string) {
	log.Fatal("service management is only supported on Windows")
}

func runWrapper(execDir string) {
	log.Fatal("wrapper mode is only supported on Windows")
}

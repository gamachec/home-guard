//go:build !windows

package main

import "log"

func isWindowsService() bool {
	return false
}

func installService() {
	log.Fatal("service management is only supported on Windows")
}

func uninstallService() {
	log.Fatal("service management is only supported on Windows")
}

func runAsService() {
	log.Fatal("service mode is only supported on Windows")
}

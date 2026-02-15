//go:build !windows

package main

import "log"

func installService() {
	log.Fatal("service management is only supported on Windows")
}

func uninstallService() {
	log.Fatal("service management is only supported on Windows")
}

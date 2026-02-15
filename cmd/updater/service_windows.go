//go:build windows

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows/svc"
)

type updaterService struct {
	execDir string
}

func (s *updaterService) Execute(_ []string, r <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("service: panic: %v", p)
		}
	}()

	log.Printf("service: starting")
	status <- svc.Status{State: svc.StartPending}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := newWrapper(s.execDir)
	go w.run(ctx)
	go w.updateLoop(ctx)

	log.Printf("service: running")
	status <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	for req := range r {
		switch req.Cmd {
		case svc.Stop, svc.Shutdown:
			log.Printf("service: stopping")
			status <- svc.Status{State: svc.StopPending}
			cancel()
			w.stopAgent()
			log.Printf("service: stopped")
			return false, 0
		}
	}
	return false, 0
}

func isWindowsService() bool {
	ok, err := svc.IsWindowsService()
	return err == nil && ok
}

func runAsService(execDir string) {
	execPath, _ := os.Executable()
	logPath := filepath.Join(filepath.Dir(execPath), "updater.log")
	if logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		defer logFile.Close()
		log.SetOutput(logFile)
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	if err := svc.Run("HomeGuard", &updaterService{execDir: execDir}); err != nil {
		log.Fatalf("service: run failed: %v", err)
	}
}

func runWrapper(execDir string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	w := newWrapper(execDir)
	go w.updateLoop(ctx)
	go w.run(ctx)

	<-sigCh
	cancel()
	w.stopAgent()
}

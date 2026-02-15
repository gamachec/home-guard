package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type wrapper struct {
	execDir string
	mu      sync.Mutex
	cmd     *exec.Cmd
	done    chan struct{}
}

func newWrapper(execDir string) *wrapper {
	return &wrapper{execDir: execDir}
}

func (w *wrapper) run(ctx context.Context) {
	for {
		agentPath := filepath.Join(w.execDir, "home-guard.exe")
		cmd := exec.CommandContext(ctx, agentPath)
		done := make(chan struct{})

		w.mu.Lock()
		w.cmd = cmd
		w.done = done
		w.mu.Unlock()

		if err := cmd.Start(); err != nil {
			log.Printf("wrapper: failed to start agent: %v", err)
			close(done)
		} else {
			go func() {
				defer close(done)
				cmd.Wait()
			}()
		}

		select {
		case <-done:
		case <-ctx.Done():
			w.killCurrent()
			<-done
			return
		}

		w.mu.Lock()
		w.cmd = nil
		w.done = nil
		w.mu.Unlock()

		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}

func (w *wrapper) killCurrent() {
	w.mu.Lock()
	cmd := w.cmd
	w.mu.Unlock()

	if cmd != nil && cmd.Process != nil {
		cmd.Process.Kill()
	}
}

func (w *wrapper) stopAgent() {
	w.mu.Lock()
	cmd := w.cmd
	done := w.done
	w.mu.Unlock()

	if cmd == nil {
		return
	}
	if cmd.Process != nil {
		cmd.Process.Kill()
	}
	if done != nil {
		<-done
	}
}

func (w *wrapper) checkAndUpdate() error {
	localVersion, err := readVersionFile(w.execDir)
	if err != nil {
		return fmt.Errorf("read version: %w", err)
	}

	release, err := latestRelease()
	if err != nil {
		return fmt.Errorf("fetch release: %w", err)
	}

	if !isNewer(release.TagName, localVersion) {
		return nil
	}

	log.Printf("updater: new version available: %s -> %s", localVersion, release.TagName)

	agentURL, checksumURL, err := findAssets(release)
	if err != nil {
		return err
	}

	newBinPath := filepath.Join(w.execDir, "home-guard.exe.new")
	if err := downloadFile(agentURL, newBinPath); err != nil {
		return fmt.Errorf("download: %w", err)
	}

	checksumsData, err := fetchBytes(checksumURL)
	if err != nil {
		os.Remove(newBinPath)
		return fmt.Errorf("download checksums: %w", err)
	}

	expectedHash, err := findChecksum(string(checksumsData), "home-guard.exe")
	if err != nil {
		os.Remove(newBinPath)
		return fmt.Errorf("parse checksums: %w", err)
	}

	if err := verifyChecksum(newBinPath, expectedHash); err != nil {
		os.Remove(newBinPath)
		return fmt.Errorf("checksum mismatch: %w", err)
	}

	agentPath := filepath.Join(w.execDir, "home-guard.exe")

	w.stopAgent()

	if err := os.Rename(newBinPath, agentPath); err != nil {
		return fmt.Errorf("replace binary: %w", err)
	}

	log.Printf("updater: updated to %s", release.TagName)
	return nil
}

func (w *wrapper) updateLoop(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(time.Minute):
	}

	if err := w.checkAndUpdate(); err != nil {
		log.Printf("updater: update check failed: %v", err)
	}

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.checkAndUpdate(); err != nil {
				log.Printf("updater: update check failed: %v", err)
			}
		}
	}
}

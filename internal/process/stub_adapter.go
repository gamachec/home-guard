//go:build !windows

package process

import "errors"

type WindowsAdapter struct{}

func NewWindowsAdapter() *WindowsAdapter {
	return &WindowsAdapter{}
}

func (a *WindowsAdapter) ListProcesses() ([]ProcessInfo, error) {
	return nil, errors.New("not supported on this platform")
}

func (a *WindowsAdapter) KillProcess(_ uint32) error {
	return errors.New("not supported on this platform")
}

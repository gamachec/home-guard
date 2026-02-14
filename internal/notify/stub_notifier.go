//go:build !windows

package notify

import "errors"

type WindowsNotifier struct{}

func NewWindowsNotifier() *WindowsNotifier {
	return &WindowsNotifier{}
}

func (n *WindowsNotifier) Send(_ Notification) error {
	return errors.New("not supported on this platform")
}

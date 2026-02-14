//go:build windows

package notify

import "github.com/go-toast/toast"

type WindowsNotifier struct{}

func NewWindowsNotifier() *WindowsNotifier {
	return &WindowsNotifier{}
}

func (n *WindowsNotifier) Send(notif Notification) error {
	t := toast.Notification{
		AppID:   "Home Guard",
		Title:   notif.Title,
		Message: notif.Message,
	}
	return t.Push()
}

package notify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type Notification struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type Notifier interface {
	Send(n Notification) error
}

func DecodeNotification(encoded string) (Notification, error) {
	payload, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return Notification{}, fmt.Errorf("base64 decode: %w", err)
	}
	var n Notification
	if err := json.Unmarshal(payload, &n); err != nil {
		return Notification{}, fmt.Errorf("json unmarshal: %w", err)
	}
	return n, nil
}

package notify

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

type mockNotifier struct {
	last Notification
}

func (m *mockNotifier) Send(n Notification) error {
	m.last = n
	return nil
}

func TestSendNotification(t *testing.T) {
	notifier := &mockNotifier{}
	n := Notification{Title: "Alerte", Message: "Temps presque écoulé"}

	if err := notifier.Send(n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if notifier.last.Title != "Alerte" {
		t.Errorf("expected title %q, got %q", "Alerte", notifier.last.Title)
	}
	if notifier.last.Message != "Temps presque écoulé" {
		t.Errorf("expected message %q, got %q", "Temps presque écoulé", notifier.last.Message)
	}
}

func TestDecodeNotification(t *testing.T) {
	original := Notification{Title: "Test", Message: "Un message"}
	payload, _ := json.Marshal(original)
	encoded := base64.StdEncoding.EncodeToString(payload)

	got, err := DecodeNotification(encoded)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Title != original.Title {
		t.Errorf("expected title %q, got %q", original.Title, got.Title)
	}
	if got.Message != original.Message {
		t.Errorf("expected message %q, got %q", original.Message, got.Message)
	}
}

func TestDecodeNotificationInvalidBase64(t *testing.T) {
	_, err := DecodeNotification("not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

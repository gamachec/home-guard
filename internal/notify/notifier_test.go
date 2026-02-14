package notify

import "testing"

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

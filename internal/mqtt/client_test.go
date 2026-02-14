package mqtt

import (
	"fmt"
	"testing"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"

	"home-guard/internal/config"
)

type mockToken struct{}

func (t *mockToken) Wait() bool                       { return true }
func (t *mockToken) WaitTimeout(d time.Duration) bool { return true }
func (t *mockToken) Done() <-chan struct{}             { ch := make(chan struct{}); close(ch); return ch }
func (t *mockToken) Error() error                     { return nil }

type mockPahoClient struct {
	isConnected bool
	published   []struct{ topic, payload string }
}

func (m *mockPahoClient) Connect() pahomqtt.Token {
	m.isConnected = true
	return &mockToken{}
}
func (m *mockPahoClient) Disconnect(quiesce uint)  { m.isConnected = false }
func (m *mockPahoClient) IsConnected() bool        { return m.isConnected }
func (m *mockPahoClient) IsConnectionOpen() bool   { return m.isConnected }
func (m *mockPahoClient) AddRoute(topic string, callback pahomqtt.MessageHandler) {}
func (m *mockPahoClient) OptionsReader() pahomqtt.ClientOptionsReader {
	return pahomqtt.ClientOptionsReader{}
}

func (m *mockPahoClient) Publish(topic string, qos byte, retained bool, payload interface{}) pahomqtt.Token {
	m.published = append(m.published, struct{ topic, payload string }{topic, fmt.Sprint(payload)})
	return &mockToken{}
}
func (m *mockPahoClient) Subscribe(topic string, qos byte, callback pahomqtt.MessageHandler) pahomqtt.Token {
	return &mockToken{}
}
func (m *mockPahoClient) SubscribeMultiple(filters map[string]byte, callback pahomqtt.MessageHandler) pahomqtt.Token {
	return &mockToken{}
}
func (m *mockPahoClient) Unsubscribe(topics ...string) pahomqtt.Token { return &mockToken{} }

func newTestClient(cfg *config.Config) (*Client, *mockPahoClient) {
	mock := &mockPahoClient{}
	client := newClientWithFactory(cfg, func(_ *pahomqtt.ClientOptions) pahomqtt.Client {
		return mock
	})
	return client, mock
}

func testConfig() *config.Config {
	return &config.Config{
		Broker:   "localhost",
		Port:     1883,
		Username: "user",
		Password: "pass",
		ClientID: "test-pc",
	}
}

func TestConnect(t *testing.T) {
	client, mock := newTestClient(testConfig())

	if err := client.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	if !mock.isConnected {
		t.Error("expected paho client to be connected")
	}
}

func TestPublishStatus(t *testing.T) {
	client, mock := newTestClient(testConfig())
	_ = client.Connect()

	if err := client.PublishStatus("online"); err != nil {
		t.Fatalf("PublishStatus() error = %v", err)
	}

	if len(mock.published) != 1 {
		t.Fatalf("expected 1 published message, got %d", len(mock.published))
	}

	expectedTopic := "stat/test-pc/status"
	if mock.published[0].topic != expectedTopic {
		t.Errorf("topic = %q, want %q", mock.published[0].topic, expectedTopic)
	}
	if mock.published[0].payload != "online" {
		t.Errorf("payload = %q, want %q", mock.published[0].payload, "online")
	}
}

func TestDisconnect(t *testing.T) {
	client, mock := newTestClient(testConfig())
	_ = client.Connect()

	client.Disconnect()

	if mock.isConnected {
		t.Error("expected paho client to be disconnected")
	}
}

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

type mockMessage struct {
	topic   string
	payload []byte
}

func (m *mockMessage) Duplicate() bool    { return false }
func (m *mockMessage) Qos() byte          { return 1 }
func (m *mockMessage) Retained() bool     { return false }
func (m *mockMessage) Topic() string      { return m.topic }
func (m *mockMessage) MessageID() uint16  { return 0 }
func (m *mockMessage) Payload() []byte    { return m.payload }
func (m *mockMessage) Ack()               {}

type mockPahoClient struct {
	isConnected      bool
	published        []struct{ topic, payload string }
	subscriptions    map[string]pahomqtt.MessageHandler
	onConnectHandler pahomqtt.OnConnectHandler
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
	if m.subscriptions == nil {
		m.subscriptions = make(map[string]pahomqtt.MessageHandler)
	}
	m.subscriptions[topic] = callback
	return &mockToken{}
}
func (m *mockPahoClient) SubscribeMultiple(filters map[string]byte, callback pahomqtt.MessageHandler) pahomqtt.Token {
	return &mockToken{}
}
func (m *mockPahoClient) Unsubscribe(topics ...string) pahomqtt.Token { return &mockToken{} }

func newTestClient(cfg *config.Config) (*Client, *mockPahoClient) {
	mock := &mockPahoClient{}
	client := newClientWithFactory(cfg, func(opts *pahomqtt.ClientOptions) pahomqtt.Client {
		mock.onConnectHandler = opts.OnConnect
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

func TestSubscribe(t *testing.T) {
	client, mock := newTestClient(testConfig())
	_ = client.Connect()

	var received string
	err := client.Subscribe("cmnd/test-pc/kill_test", func(payload []byte) {
		received = string(payload)
	})
	if err != nil {
		t.Fatalf("Subscribe() error = %v", err)
	}

	handler, ok := mock.subscriptions["cmnd/test-pc/kill_test"]
	if !ok {
		t.Fatal("expected subscription to be registered on topic")
	}

	handler(mock, &mockMessage{topic: "cmnd/test-pc/kill_test", payload: []byte("roblox.exe")})
	if received != "roblox.exe" {
		t.Errorf("received = %q, want %q", received, "roblox.exe")
	}
}

func TestPublish(t *testing.T) {
	client, mock := newTestClient(testConfig())
	_ = client.Connect()

	if err := client.Publish("stat/test-pc/current_mode", "BLOCKED"); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	if len(mock.published) != 1 {
		t.Fatalf("expected 1 published message, got %d", len(mock.published))
	}
	if mock.published[0].topic != "stat/test-pc/current_mode" {
		t.Errorf("topic = %q, want %q", mock.published[0].topic, "stat/test-pc/current_mode")
	}
	if mock.published[0].payload != "BLOCKED" {
		t.Errorf("payload = %q, want %q", mock.published[0].payload, "BLOCKED")
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

func TestOnConnectCallback(t *testing.T) {
	client, mock := newTestClient(testConfig())

	called := false
	client.SetOnConnect(func() { called = true })

	_ = client.Connect()

	if mock.onConnectHandler == nil {
		t.Fatal("expected OnConnect handler to be registered")
	}

	mock.onConnectHandler(mock)

	if !called {
		t.Error("expected OnConnect callback to be called")
	}
}

func TestPublishDiscovery(t *testing.T) {
	client, mock := newTestClient(testConfig())
	_ = client.Connect()

	if err := client.PublishDiscovery(); err != nil {
		t.Fatalf("PublishDiscovery() error = %v", err)
	}

	expectedTopics := []string{
		"homeassistant/select/test-pc/mode/config",
		"homeassistant/binary_sensor/test-pc/connectivity/config",
		"homeassistant/sensor/test-pc/apps/config",
		"homeassistant/sensor/test-pc/version/config",
	}

	if len(mock.published) != len(expectedTopics) {
		t.Fatalf("expected %d published messages, got %d", len(expectedTopics), len(mock.published))
	}

	for i, topic := range expectedTopics {
		if mock.published[i].topic != topic {
			t.Errorf("published[%d].topic = %q, want %q", i, mock.published[i].topic, topic)
		}
	}
}

func TestPublishVersion(t *testing.T) {
	client, mock := newTestClient(testConfig())
	_ = client.Connect()

	if err := client.PublishVersion("v1.2.3"); err != nil {
		t.Fatalf("PublishVersion() error = %v", err)
	}

	if len(mock.published) != 1 {
		t.Fatalf("expected 1 published message, got %d", len(mock.published))
	}
	if mock.published[0].topic != "stat/test-pc/version" {
		t.Errorf("topic = %q, want %q", mock.published[0].topic, "stat/test-pc/version")
	}
	if mock.published[0].payload != "v1.2.3" {
		t.Errorf("payload = %q, want %q", mock.published[0].payload, "v1.2.3")
	}
}

func TestExponentialDelay(t *testing.T) {
	cases := []struct {
		attempt int
		want    time.Duration
	}{
		{1, time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
		{11, 2 * time.Minute},
	}

	for _, tc := range cases {
		got := exponentialDelay(tc.attempt)
		if got != tc.want {
			t.Errorf("exponentialDelay(%d) = %s, want %s", tc.attempt, got, tc.want)
		}
	}
}

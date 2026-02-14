package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"

	"home-guard/internal/config"
)

const (
	connectTimeout = 10 * time.Second
	baseDelay      = time.Second
	maxDelay       = 2 * time.Minute
	maxRetries     = 12
)

type pahoFactory func(*pahomqtt.ClientOptions) pahomqtt.Client

type Client struct {
	cfg       *config.Config
	paho      pahomqtt.Client
	factory   pahoFactory
	onConnect func()
}

func NewClient(cfg *config.Config) *Client {
	return newClientWithFactory(cfg, pahomqtt.NewClient)
}

func newClientWithFactory(cfg *config.Config, factory pahoFactory) *Client {
	return &Client{cfg: cfg, factory: factory}
}

func (c *Client) SetOnConnect(fn func()) {
	c.onConnect = fn
}

func (c *Client) Connect() error {
	opts := c.buildOptions()
	c.paho = c.factory(opts)

	for attempt := 0; ; attempt++ {
		if attempt > 0 {
			delay := exponentialDelay(attempt)
			log.Printf("MQTT reconnect attempt %d in %s", attempt+1, delay)
			time.Sleep(delay)
		}

		token := c.paho.Connect()
		if !token.WaitTimeout(connectTimeout) {
			log.Printf("MQTT connect timeout (attempt %d)", attempt+1)
		} else if err := token.Error(); err != nil {
			log.Printf("MQTT connect error (attempt %d): %v", attempt+1, err)
		} else {
			return nil
		}

		if attempt >= maxRetries {
			return fmt.Errorf("failed to connect to MQTT broker after %d attempts", maxRetries+1)
		}
	}
}

func exponentialDelay(attempt int) time.Duration {
	delay := float64(baseDelay) * math.Pow(2, float64(attempt-1))
	if delay > float64(maxDelay) {
		return maxDelay
	}
	return time.Duration(delay)
}

func (c *Client) PublishStatus(status string) error {
	topic := fmt.Sprintf("stat/%s/status", c.cfg.ClientID)
	token := c.paho.Publish(topic, 1, true, status)
	token.Wait()
	return token.Error()
}

func (c *Client) Publish(topic string, payload string) error {
	token := c.paho.Publish(topic, 1, true, payload)
	token.Wait()
	return token.Error()
}

type haDevice struct {
	Identifiers  []string `json:"identifiers"`
	Name         string   `json:"name,omitempty"`
	Model        string   `json:"model,omitempty"`
	Manufacturer string   `json:"manufacturer,omitempty"`
}

type haSelectDiscovery struct {
	Name         string   `json:"name"`
	UniqueID     string   `json:"unique_id"`
	CommandTopic string   `json:"command_topic"`
	StateTopic   string   `json:"state_topic"`
	Options      []string `json:"options"`
	Device       haDevice `json:"device"`
}

type haBinarySensorDiscovery struct {
	Name        string   `json:"name"`
	UniqueID    string   `json:"unique_id"`
	DeviceClass string   `json:"device_class"`
	StateTopic  string   `json:"state_topic"`
	PayloadOn   string   `json:"payload_on"`
	PayloadOff  string   `json:"payload_off"`
	Device      haDevice `json:"device"`
}

type haSensorDiscovery struct {
	Name       string   `json:"name"`
	UniqueID   string   `json:"unique_id"`
	StateTopic string   `json:"state_topic"`
	Device     haDevice `json:"device"`
}

func (c *Client) PublishDiscovery() error {
	id := c.cfg.ClientID
	fullDevice := haDevice{
		Identifiers:  []string{id},
		Name:         id,
		Model:        "HomeGuard",
		Manufacturer: "HomeGuard",
	}
	minDevice := haDevice{Identifiers: []string{id}}

	entries := []struct {
		topic   string
		payload any
	}{
		{
			fmt.Sprintf("homeassistant/select/%s/mode/config", id),
			haSelectDiscovery{
				Name:         "Mode d'utilisation",
				UniqueID:     id + "_mode",
				CommandTopic: fmt.Sprintf("cmnd/%s/mode", id),
				StateTopic:   fmt.Sprintf("stat/%s/current_mode", id),
				Options:      []string{"ACTIVE", "BLOCKED"},
				Device:       fullDevice,
			},
		},
		{
			fmt.Sprintf("homeassistant/binary_sensor/%s/connectivity/config", id),
			haBinarySensorDiscovery{
				Name:        "Etat",
				UniqueID:    id + "_online",
				DeviceClass: "connectivity",
				StateTopic:  fmt.Sprintf("stat/%s/status", id),
				PayloadOn:   "online",
				PayloadOff:  "offline",
				Device:      minDevice,
			},
		},
		{
			fmt.Sprintf("homeassistant/sensor/%s/apps/config", id),
			haSensorDiscovery{
				Name:       "Applications en cours",
				UniqueID:   id + "_running_apps",
				StateTopic: fmt.Sprintf("stat/%s/running_apps", id),
				Device:     minDevice,
			},
		},
	}

	for _, e := range entries {
		data, err := json.Marshal(e.payload)
		if err != nil {
			return err
		}
		token := c.paho.Publish(e.topic, 1, true, data)
		token.Wait()
		if err := token.Error(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) PublishRunningApps(apps []string) error {
	topic := fmt.Sprintf("stat/%s/running_apps", c.cfg.ClientID)
	payload, err := json.Marshal(apps)
	if err != nil {
		return err
	}
	token := c.paho.Publish(topic, 1, true, payload)
	token.Wait()
	return token.Error()
}

func (c *Client) Subscribe(topic string, handler func(payload []byte)) error {
	token := c.paho.Subscribe(topic, 1, func(_ pahomqtt.Client, msg pahomqtt.Message) {
		handler(msg.Payload())
	})
	token.Wait()
	return token.Error()
}

func (c *Client) Disconnect() {
	if c.paho != nil && c.paho.IsConnected() {
		c.paho.Disconnect(250)
	}
}

func (c *Client) buildOptions() *pahomqtt.ClientOptions {
	broker := fmt.Sprintf("tcp://%s:%d", c.cfg.Broker, c.cfg.Port)
	lwtTopic := fmt.Sprintf("stat/%s/status", c.cfg.ClientID)

	return pahomqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(c.cfg.ClientID).
		SetUsername(c.cfg.Username).
		SetPassword(c.cfg.Password).
		SetWill(lwtTopic, "offline", 1, true).
		SetAutoReconnect(true).
		SetCleanSession(false).
		SetOnConnectHandler(func(_ pahomqtt.Client) {
			if c.onConnect != nil {
				c.onConnect()
			}
		})
}

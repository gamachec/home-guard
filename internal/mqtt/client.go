package mqtt

import (
	"fmt"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"

	"home-guard/internal/config"
)

type pahoFactory func(*pahomqtt.ClientOptions) pahomqtt.Client

type Client struct {
	cfg     *config.Config
	paho    pahomqtt.Client
	factory pahoFactory
}

func NewClient(cfg *config.Config) *Client {
	return newClientWithFactory(cfg, pahomqtt.NewClient)
}

func newClientWithFactory(cfg *config.Config, factory pahoFactory) *Client {
	return &Client{cfg: cfg, factory: factory}
}

func (c *Client) Connect() error {
	opts := c.buildOptions()
	c.paho = c.factory(opts)

	token := c.paho.Connect()
	if !token.WaitTimeout(10 * time.Second) {
		return fmt.Errorf("connection timeout")
	}
	return token.Error()
}

func (c *Client) PublishStatus(status string) error {
	topic := fmt.Sprintf("stat/%s/status", c.cfg.ClientID)
	token := c.paho.Publish(topic, 1, true, status)
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
		SetCleanSession(false)
}

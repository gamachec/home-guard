package mqtt

import (
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

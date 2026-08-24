// Package mqtt publishes quota payloads to an MQTT broker.
package mqtt

import (
	"fmt"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

// Publisher publishes retained messages to an MQTT topic.
type Publisher struct {
	client paho.Client
	topic  string
}

// NewPublisher connects to the broker and returns a Publisher for topic.
func NewPublisher(broker, clientID, topic, username, password string) (*Publisher, error) {
	opts := paho.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second).
		SetConnectTimeout(10 * time.Second)
	if len(username) > 0 {
		opts.SetUsername(username)
		opts.SetPassword(password)
	}

	client := paho.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(15 * time.Second) {
		return nil, fmt.Errorf("connect to %s: timeout", broker)
	}
	if err := token.Error(); err != nil {
		return nil, fmt.Errorf("connect to %s: %w", broker, err)
	}
	return &Publisher{client: client, topic: topic}, nil
}

// Publish sends payload to the topic as a retained QoS 1 message.
func (p *Publisher) Publish(payload []byte) error {
	token := p.client.Publish(p.topic, 1, true, payload)
	if !token.WaitTimeout(10 * time.Second) {
		return fmt.Errorf("publish to %s: timeout", p.topic)
	}
	return token.Error()
}

// Close disconnects from the broker.
func (p *Publisher) Close() {
	p.client.Disconnect(1000)
}

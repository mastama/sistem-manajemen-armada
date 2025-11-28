package mqttclient

import (
	"log"
	"time"

	"sistem-manajemen-armada/internal/config"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func NewClient(cfg *config.Config) mqtt.Client {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTTBrokerURL).
		SetClientID(cfg.MQTTClientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(3 * time.Second)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("failed to connect MQTT broker: %v", token.Error())
	}
	log.Println("Connected to MQTT broker")

	return client
}

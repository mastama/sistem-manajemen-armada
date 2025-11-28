package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"sistem-manajemen-armada/internal/config"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	cfg := config.Load()

	opts := mqtt.NewClientOptions().
		AddBroker(cfg.MQTTBrokerURL).
		SetClientID("fleet-mock-publisher")

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("mqtt connect error: %v", token.Error())
	}
	defer client.Disconnect(250)

	vehicleID := "B1234XYZ"
	topic := "/fleet/vehicle/" + vehicleID + "/location"

	rand.Seed(time.Now().UnixNano())
	log.Println("Mock publisher started, topic:", topic)

	for {
		lat := cfg.GeofenceLat + (rand.Float64()-0.5)/1000
		lon := cfg.GeofenceLon + (rand.Float64()-0.5)/1000

		payload := map[string]any{
			"vehicle_id": vehicleID,
			"latitude":   lat,
			"longitude":  lon,
			"timestamp":  time.Now().Unix(),
		}

		b, _ := json.Marshal(payload)
		token := client.Publish(topic, 0, false, b)
		token.Wait()

		log.Printf("Published mock location: %s", string(b))
		time.Sleep(2 * time.Second)
	}
}

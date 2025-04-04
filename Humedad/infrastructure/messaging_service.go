package infrastructure

import (
	"encoding/json"
	"log"
	"time"

	"github.com/streadway/amqp"
)

type MessagingService struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	hub  *Hub
}

func NewMessagingService(hub *Hub) *MessagingService {
	conn, err := amqp.Dial("amqp://reyhades:reyhades@44.223.218.9:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
		return nil
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
		return nil
	}

	err = ch.ExchangeDeclare(
		"amq.topic", // exchange name
		"topic",     // type
		true,        // durable
		false,       // auto-deleted
		false,       // internal
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare an exchange: %s", err)
		return nil
	}

	return &MessagingService{
		conn: conn,
		ch:   ch,
		hub:  hub,
	}
}

func (ms *MessagingService) ConsumeTemperatureMessages() error {
	q, err := ms.ch.QueueDeclare(
		"esp32humedad", // queue name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return err
	}

	err = ms.ch.QueueBind(
		q.Name,         // queue name
		"sp32.humedad", // routing key
		"amq.topic",    // exchange
		false,
		nil,
	)
	if err != nil {
		return err
	}

	msgs, err := ms.ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack (false for manual acknowledgment)
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			var rawData map[string]interface{}
			if err := json.Unmarshal(msg.Body, &rawData); err != nil {
				log.Printf("Error parsing JSON: %v", err)
				msg.Nack(false, true) // Requeue the message
				continue
			}

			// Transform data to match frontend interface
			humedadValue, ok := rawData["value"].(float64)
			if !ok {
				log.Printf("Invalid humidity value: %v", rawData["value"])
				msg.Ack(false)
				continue
			}

			adaptedData := map[string]interface{}{
				"humedad":     humedadValue,
				"unit":        rawData["unit"],
				"device":      rawData["device_id"],
				"ts":          rawData["timestamp"],
				"created_at":  time.Now().Format(time.RFC3339),
				"sensor_type": rawData["sensor_type"],
			}

			adaptedMessage, err := json.Marshal(adaptedData)
			if err != nil {
				log.Printf("Error marshaling adapted data: %v", err)
				msg.Ack(false)
				continue
			}

			log.Printf("Sending to WebSocket: %s", string(adaptedMessage))
			ms.hub.broadcast <- adaptedMessage
			msg.Ack(false)
		}
	}()

	return nil
}

func (ms *MessagingService) Close() {
	if ms.ch != nil {
		ms.ch.Close()
	}
	if ms.conn != nil {
		ms.conn.Close()
	}
}

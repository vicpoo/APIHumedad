// messaging_service.go
package infrastructure

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
)

type MessagingService struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	hub  *Hub
}

func NewMessagingService(hub *Hub) *MessagingService {
	// Note: Port 1883 is for MQTT, RabbitMQ typically uses 5672 for AMQP
	// If you really need to use 1883, you might need an MQTT plugin in RabbitMQ
	// or use an MQTT client library instead of AMQP
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

	// Declare the exchange (if it doesn't exist)
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
	// Declare the queue (if it doesn't exist)
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

	// Bind the queue to the exchange with the correct routing key
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
			// Log the raw message
			log.Printf("Received raw message: %s", string(msg.Body))

			// Parse the JSON to make it more readable
			var data map[string]interface{}
			if err := json.Unmarshal(msg.Body, &data); err == nil {
				log.Printf("Parsed temperature data: %+v", data)
			} else {
				log.Printf("Error parsing JSON: %v", err)
			}

			// Send the message to all WebSocket clients
			ms.hub.broadcast <- msg.Body

			// Acknowledge successful processing
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

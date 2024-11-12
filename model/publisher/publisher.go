package publisher

import (
	"context"
	"os"
	// "fmt"
	// "os"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

//go:generate mockgen -source=publisher.go -destination=mock_publisher/mock_publisher.go -package=mock_publisher

type publisher struct {
	conn *amqp.Connection
}

type Publisher interface {
	PublishDefaultExchange(ctx context.Context, body []byte) error
}

func NewPublisher() Publisher {
	return &publisher{}
}

func (p *publisher) newConnection() error {
	amqpURL := os.Getenv("AMQP_URL")

	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return err
	}
	p.conn = conn
	return nil
}

func (p *publisher) ensureConnection() error {
	if p.conn == nil {
		return p.newConnection()
	}

	ch, err := p.conn.Channel()
	if err != nil {
		log.Fatalln("Unable to connect to rabbitMQ:", err)
		return p.newConnection()
	}

	ch.Close()
	return nil
}

func (p *publisher) PublishDefaultExchange(ctx context.Context, reportJSON []byte) error {
	err := p.ensureConnection()
	if err != nil {
		return err
	}

	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"Receiver", // exchange name
		"topic",    // exchange type
		true,       // durable
		false,      // auto-delete
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	)
	failOnError(err, "Failed to declare a queue")

	routingKey := "reservation.created"
	err = ch.Publish(
		"Receiver", // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        reportJSON,
		},
	)
	if err != nil {
		log.Fatalf("Failed to publish message: %s", err)
	}

	log.Println("Published message")
	return nil
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
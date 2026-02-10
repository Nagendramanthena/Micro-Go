package event

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type emitter struct {
	connection *amqp.Connection
}

func (e *emitter) setup() error {
	channel, err := e.connection.Channel()

	if err != nil {
		return err
	}
	defer channel.Close()
	return declareExchange(channel)

}

func (e *emitter) Push(event string, severity string) error {
	channel, err := e.connection.Channel()

	if err != nil {
		return err
	}
	defer channel.Close()
	log.Println("Pushing to channel")

	err = channel.Publish(
		"logs_topic",
		severity,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(event),
		},
	)

	if err != nil {
		return nil
	}

	return nil

}

func NewEventEmitter(conn *amqp.Connection) (emitter, error) {
	emit := emitter{
		connection: conn,
	}

	err := emit.setup()

	if err != nil {
		return emitter{}, err
	}

	return emit, nil
}

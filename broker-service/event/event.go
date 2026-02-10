package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func declareExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		"logs_topic", // Name of the channel
		"topic",      // type
		true,         //durable?
		false,        // auto-deleted?
		false,        //inertnal ??
		false,        // no-wait?
		nil,          // arguments??
	)
}

func declareRandomQueue(ch *amqp.Channel) (amqp.Queue, error) {
	return ch.QueueDeclare(
		"",    //name?
		false, //durable??
		false, //auto-delete?
		true,  //exclusive
		false, //no-wait
		nil,   //no arguments
	)
}

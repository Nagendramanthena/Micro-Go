package main

import (
	"fmt"
	"listener/event"
	"log"
	"math"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	//try to connect with rabbitmq

	rabbitConn, err := connect()

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	defer rabbitConn.Close()

	// start listening for messages
	log.Printf("started listening to push to rabbitMq")

	//create a consumer
	consumer, err := event.NewConsumer(rabbitConn)

	if err != nil {
		panic(err)
	}

	//watch the queue and consume the events
	err = consumer.Listen([]string{"log.INFO", "log.Warning", "log.ERROR"})
	if err != nil {
		panic(err)
	}

	log.Printf("closing off the connection as we are out of this function")
}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backoff = time.Second * 1
	var connection *amqp.Connection

	for {
		c, err := amqp.Dial("amqp://admin:admin@rabbitmq")

		if err != nil {
			fmt.Println("RabbitMq not yet ready")
			counts++
		} else {
			connection = c
			log.Printf("connected to RabbitMq")
			break
		}

		if counts > 5 {
			fmt.Println(err)
			return nil, err
		}

		backoff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Printf("backing off  -- waiting for rabbitMq")
		time.Sleep(backoff)
		continue
	}

	return connection, nil

}

package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webPort = "80"

type config struct {
	Rabbit *amqp.Connection
}

func main() {
	rabbitConn, err := connect()

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	defer rabbitConn.Close()
	app := config{
		Rabbit: rabbitConn,
	}

	log.Printf("Starting Broker service on port %s\n", webPort)

	//define HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	//START THE SERVER
	err = srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
		// This is a placeholder for the API service main function.
	}
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

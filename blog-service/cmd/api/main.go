package main

import (
	"blog/data"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	webport  = "80"
	mongoUrl = "mongodb://admin:password@mongo-blogs:27017/blogs_db?authSource=admin"
)

var client *mongo.Client

type config struct {
	Models    data.Models
	redclinet *redis.Client
}

func main() {
	//connect with redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",      // NOT localhost
		Password: "strong-password", // No password
		DB:       0,
	})

	ctx := context.Background()

	pong, err := redisClient.Ping(ctx).Result()

	if err != nil {
		log.Printf("not able to connect with redis from main.go file")

	}

	log.Printf("Pong error", pong)

	//connect to mongo
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}

	client = mongoClient

	//close connection

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Panic(err)
		}
	}()

	app := config{
		Models:    data.New(client),
		redclinet: redisClient,
	}

	//start web server

	log.Printf("Starting logger service on port %s", webport)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webport),
		Handler: app.routes(),
	}

	er := srv.ListenAndServe()
	if er != nil {
		log.Panic(err)
	}

}

// func (app *Config) serve() {
// 	srv := &http.Server{
// 		Addr:    fmt.Sprintf(":%s", webport),
// 		Handler: app.routes(),
// 	}

// 	err := srv.ListenAndServe()
// 	if err != nil {
// 		log.Panic(err)
// 	}
// }

func connectToMongo() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(mongoUrl)

	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	//connect

	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("Error connecting to mongo:", err)
		return nil, err
	}

	log.Printf("Connected to MongoDB at %s", mongoUrl)

	return c, nil

}

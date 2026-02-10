package data

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func New(mongo *mongo.Client) Models {
	client = mongo

	return Models{
		BlogEntry: BlogEntry{},
	}

}

type Models struct {
	BlogEntry BlogEntry
}

type BlogEntry struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string    `bson:"name" json:"name"`
	Data      string    `bson:"data" json:"data"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

func (l *BlogEntry) Insert(entry BlogEntry) error {
	collection := client.Database("logs").Collection("logs")
	log.Println("Inserting log entry:", entry)
	_, err := collection.InsertOne(context.TODO(), BlogEntry{
		Name:      entry.Name,
		Data:      entry.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err != nil {
		log.Println("Error inserting log entry:", err)
		return err
	}

	return nil

}

func (l *BlogEntry) All() ([]*BlogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

	opts := options.Find()
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(context.TODO(), bson.D{}, opts)

	if err != nil {
		log.Println("Error Finding all log entries:", err)
		return nil, err
	}

	defer cursor.Close(ctx)

	var logs []*BlogEntry

	for cursor.Next(ctx) {
		var item BlogEntry
		err := cursor.Decode(&item)
		if err != nil {
			log.Println("Error decoding log into Slice:", err)
			return nil, err
		} else {
			logs = append(logs, &item)
		}
	}

	return logs, nil

}

func (l *BlogEntry) GetOne(id string) (*BlogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer cancel()
	collection := client.Database("logs").Collection("logs")

	docId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Error converting id to ObjectID:", err)
		return nil, err
	}

	var logEntry BlogEntry

	err = collection.FindOne(ctx, bson.M{"_id": docId}).Decode(&logEntry)
	if err != nil {
		log.Println("Error getting one log entry:", err)
		return nil, err
	}

	return &logEntry, nil
}

func (l *BlogEntry) DropCollection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

	err := collection.Drop(ctx)
	if err != nil {
		log.Println("Error dropping collection:", err)
		return err
	}

	return nil
}

func (l *BlogEntry) Update() (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

	docId, err := primitive.ObjectIDFromHex(l.ID)
	if err != nil {
		return nil, err
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": docId}, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "name", Value: l.Name},
			{Key: "data", Value: l.Data},
			{Key: "updated_at", Value: time.Now()},
		}},
	})

	if err != nil {
		log.Println("Error updating log entry:", err)
		return nil, err
	}

	return result, nil
}

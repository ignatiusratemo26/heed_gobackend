package main

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Transcription struct {
	Title   string    `bson:"title"`
	Content string    `bson:"content"`
	Created time.Time `bson:"created"`
}

func ConnectMongoDB() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI("mongodb://localhot:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}
	err = client.Ping(context.TODO(), nil)
	return client, err
}

func StoreTranscription(client *mongo.Client, title string, content string) error {
	collection := client.Database("transcriptionDB").Collection("transcriptions")
	transcription := Transcription{
		Title:   title,
		Content: content,
		Created: time.Now(),
	}
	_, err := collection.InsertOne(context.TODO(), transcription)
	return err
}

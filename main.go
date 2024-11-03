package main

import (
	"context"
	"log"
	"net/http"
)

func main() {
	client, err := ConnectMongoDB()
	//connecting to MongoDB
	if err != nil {
		log.Fatal("Error connecting to MongoDB: ", err)
	}
	defer client.Disconnect(context.TODO())

	//api routes
	http.HandleFunc("/api/record", StartRecordingHandler)
	http.HandleFunc("/api/transcribe", TranscribeAndStoreHandler(client))

	log.Println("Server starting on port 8080...")
	http.ListenAndServe(":8080", nil)

}

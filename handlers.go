package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

type TranscriptionRequest struct {
	Title string `json:"title"`
}

func StartRecordingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Recording started; client should manage recording and upload audio file."))

}

func TranscribeAudio(filePath string) (string, error) {

	apiURL := "https://eastus.api.cognitive.microsoft.com/"
	apiKey := os.Getenv("AZURE_API_KEY")

	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	part, _ := writer.CreateFormFile("file", filePath)
	io.Copy(part, file)
	writer.Close()

	req, err := http.NewRequest("POST", apiURL, &requestBody)
	if err != nil {
		return "", errors.New("error in sennding transcription")
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.NewRequest("POST", apiURL, &requestBody)
	if err != nil || resp.Response.StatusCode != http.StatusOK {
		return "", errors.New("error in transcription request")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
}

func TranscribeAndStoreHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//parsing title from the request boby
		var request TranscriptionRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil || request.Title == "" {
			http.Error(w, "Invalid title provided", http.StatusBadRequest)
			return
		}

		// read audio file from the form-data
		file, _, err := r.FormFile("audio")
		if err != nil {
			http.Error(w, "Error reading audio file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		//savid audio file locally
		audioPath := "temp_audio.wav"
		outFile, err := os.Create(audioPath)
		if err != nil {
			http.Error(w, "Could not save the audio file", http.StatusInternalServerError)
			return
		}
		defer outFile.Close()
		io.Copy(outFile, file)

		//sending audio to transcription service
		transcriptionText, err := TranscribeAudio(audioPath)
		if err != nil {
			http.Error(w, "Transcription failed", http.StatusInternalServerError)
			return
		}
		// storing in mongo db
		err = StoreTranscription(client, request.Title, transcriptionText)
		if err != nil {
			http.Error(w, "Failed to store transcription", http.StatusInternalServerError)
			return
		}

		//return success response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Transcription saved with title: %s", request.Title)))

	}
}

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

func init() {
	functions.CloudEvent("UploadVideo", uploadVideo)
}

type StorageObjectData struct {
	Bucket string `json:"bucket,omitempty"`
	Name   string `json:"name,omitempty"`
}

func uploadVideo(ctx context.Context, event event.Event) error {
	// Retrieve the bucket and object information from the event
	var data StorageObjectData
	if err := event.DataAs(&data); err != nil {
		log.Printf("Error parsing CloudEvent data: %v", err)
		return fmt.Errorf("event.DataAs: %v", err)
	}

	// Create a new Google Cloud Storage client
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Printf("Error creating storage client: %v", err)
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	// Download the video file from Cloud Storage
	compilationObject := client.Bucket(data.Bucket).Object(data.Name)
	compilationFilePath := fmt.Sprintf("/tmp/%s", data.Name)

	compilationFile, err := os.Create(compilationFilePath)
	if err != nil {
		log.Printf("Error creating compilation file: %v", err)
		return fmt.Errorf("os.Create: %v", err)
	}
	defer compilationFile.Close()

	reader, err := compilationObject.NewReader(ctx)
	if err != nil {
		log.Printf("Error reading compilation object: %v", err)
		return fmt.Errorf("compilationObject.NewReader: %v", err)
	}
	defer reader.Close()

	if _, err := io.Copy(compilationFile, reader); err != nil {
		log.Printf("Error downloading compilation file: %v", err)
		return fmt.Errorf("io.Copy: %v", err)
	}

	// Create a new YouTube service client
	youtubeService, err := youtube.NewService(ctx, option.WithCredentialsFile("./creds.json"))
	if err != nil {
		return fmt.Errorf("failed to create YouTube client: %v", err)
	}

	// Create a new video upload
	video := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       "Video Title",
			Description: "Video Description",
			CategoryId:  "22", // Replace with the appropriate category ID
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: "private", // Set the privacy status of the video
		},
	}

	// Create the video insert call
	call := youtubeService.Videos.Insert([]string{"snippet", "status"}, video)

	// Reset the position of the compilation file
	if _, err := compilationFile.Seek(0, 0); err != nil {
		log.Printf("Error resetting compilation file position: %v", err)
		return fmt.Errorf("compilationFile.Seek: %v", err)
	}

	// Set the video file as the media content
	call.Media(compilationFile)

	// Execute the video upload
	_, err = call.Do()
	if err != nil {
		return fmt.Errorf("failed to upload video: %v", err)
	}

	log.Printf("Video uploaded successfully: %s", data.Name)
	return nil
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"cloud.google.com/go/storage"
)

const (
	bucketName = "videos-quarantine-2486aa1dcdb442fda0c2f090761b4479"
)

type RequestBody struct {
	URL string `json:"url"`
}

func main() {
	log.Print("starting server...")
	http.HandleFunc("/", handler)

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	// Start HTTP server.
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse the JSON request body
	var requestBody RequestBody
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		http.Error(w, "Failed to parse JSON request body", http.StatusBadRequest)
		return
	}

	// Get the value of the "url" field from the request body
	videoURL := requestBody.URL
	if videoURL == "" {
		http.Error(w, "Missing 'url' field in request body", http.StatusBadRequest)
		return
	}

	// Set the path to the "yt-dlp" binary
	ytdlpPath := "/usr/local/bin/yt-dlp"

	// Set the additional flags for "yt-dlp"
	format := "bv*[ext=mp4]+ba[ext=m4a]/b[ext=mp4]"
	outputTemplate := "%(extractor)s-%(id)s.%(ext)s"
	targetDir := "/tmp"
	videoFileTemplate := filepath.Join(targetDir, outputTemplate)

	// Create the "yt-dlp" command with the specified flags
	cmd := exec.Command(ytdlpPath, "--format", format, "-o", videoFileTemplate, "--restrict-filenames", videoURL)

	// Execute the "yt-dlp" command to download the video
	output, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error downloading video: %s", string(output)), http.StatusInternalServerError)
		return
	}

	log.Printf("yt-dlp output: %s", string(output))

	// Find the downloaded MP4 file in the "/tmp" directory
	files, err := filepath.Glob("/tmp/*.mp4")
	if err != nil {
		http.Error(w, "Failed to search for downloaded video file", http.StatusInternalServerError)
		return
	}

	if len(files) == 0 {
		http.Error(w, "No downloaded video file found", http.StatusInternalServerError)
		return
	}

	videoFilePath := files[0]

	// Create a new Cloud Storage client
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		http.Error(w, "Failed to create Cloud Storage client", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// Create a new object handle in the bucket
	obj := client.Bucket(bucketName).Object(filepath.Base(videoFilePath))

	// Check if the video file already exists in the bucket
	exists, err := obj.Attrs(ctx)
	if err == nil && exists != nil {
		// Video file already exists in the bucket
		fmt.Fprintf(w, "Video file already exists in the bucket: %s", obj.ObjectName())

		// Delete the temporary video file from the container
		err = os.Remove(videoFilePath)
		if err != nil {
			log.Printf("Failed to delete temporary video file: %s", err)
		}

		return
	}

	// Open the downloaded video file
	videoFile, err := os.Open(videoFilePath)
	if err != nil {
		http.Error(w, "Failed to open video file", http.StatusInternalServerError)
		return
	}
	defer videoFile.Close()

	// Upload the video file to Cloud Storage
	writer := obj.NewWriter(ctx)
	if _, err := io.Copy(writer, videoFile); err != nil {
		http.Error(w, "Failed to upload video to Cloud Storage", http.StatusInternalServerError)
		return
	}
	if err := writer.Close(); err != nil {
		http.Error(w, "Failed to close Cloud Storage writer", http.StatusInternalServerError)
		return
	}

	// Delete the temporary video file from the container
	err = os.Remove(videoFilePath)
	if err != nil {
		log.Printf("Failed to delete temporary video file: %s", err)
	}

	// Send a response back to the client
	fmt.Fprintf(w, "Video downloaded and uploaded to Cloud Storage bucket: %s", bucketName)
}

package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"cloud.google.com/go/storage"
)

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
	// Parse the request parameters
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse request parameters", http.StatusBadRequest)
		return
	}

	// Get the value of the "url" parameter
	videoURL := r.FormValue("url")
	if videoURL == "" {
		http.Error(w, "Missing 'url' parameter", http.StatusBadRequest)
		return
	}

	// Send a response back to the client
	fmt.Fprintf(w, "Video downloading and uploading process started")

	// Start a goroutine to handle the downloading and uploading process
	go func() {

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

		// Set the name of the Cloud Storage bucket
		bucketName := "videos-quarantine-2486aa1dcdb442fda0c2f090761b4479"

		// Open the downloaded video file
		videoFile, err := os.Open(videoFilePath)
		if err != nil {
			http.Error(w, "Failed to open video file", http.StatusInternalServerError)
			return
		}
		defer videoFile.Close()

		// Create a new object handle in the bucket
		obj := client.Bucket(bucketName).Object(filepath.Base(videoFilePath))

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

		// Send a response back to the client
		fmt.Fprintf(w, "Video downloaded and uploaded to Cloud Storage bucket: %s", bucketName)

	}()
}

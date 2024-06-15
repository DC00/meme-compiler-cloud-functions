package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	normalizedVideoBucket = "gs://videos-normalized-3ec32eeafcfe42f28cb86296afa48673"
	compilationsBucket    = "gs://compilations-f714ffc72eaf414ea0f51b18f4678383"
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
	// Create a temporary directory to store the downloaded videos
	tempDir, err := os.MkdirTemp("", "normalized-videos")
	if err != nil {
		log.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Download the videos from the "normalized" bucket using gsutil
	cmd := exec.Command("gsutil", "-m", "cp", normalizedVideoBucket+"/*", tempDir)
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to download videos: %v", err)
	}

	// Get the list of downloaded video files
	videoFiles, err := filepath.Glob(filepath.Join(tempDir, "*"))
	if err != nil {
		log.Fatalf("Failed to get video files: %v", err)
	}

	// Create the video list file for ffmpeg
	videoListFile := filepath.Join(tempDir, "videos-for-ffmpeg.txt")
	file, err := os.Create(videoListFile)
	if err != nil {
		log.Fatalf("Failed to create video list file: %v", err)
	}
	defer file.Close()

	for _, videoFile := range videoFiles {
		fmt.Fprintf(file, "file '%s'\n", videoFile)
	}

	// Run ffmpeg command to stitch the videos together
	outputFile := filepath.Join(tempDir, "output.mp4")
	cmd = exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", videoListFile, "-c", "copy", outputFile)

	// If video quality is poor we may need to reencode
	/*
		cmd := exec.Command("ffmpeg",
			"-f", "concat",
			"-safe", "0",
			"-i", videoListFile,
			"-c:v", "libx264",
			"-preset", "veryslow",
			"-crf", "21",
			"-pix_fmt", "yuv420p",
			"-c:a", "aac",
			"-ar", "48000",
			"-b:a", "384k",
			outputFile,
		)
	*/

	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run ffmpeg command: %v", err)
	}

	// Upload the compilation video to the "compilation" bucket using gsutil
	cmd = exec.Command("gsutil", "cp", outputFile, compilationsBucket)
	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to upload compilation video: %v", err)
	}

	// Delete the normalized videos from the "normalized" bucket using gsutil
	cmd = exec.Command("gsutil", "-m", "rm", normalizedVideoBucket+"/*")
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to delete normalized videos: %v", err)
	}

	fmt.Fprintf(w, "Compilation video created and uploaded successfully.")
}

package concatenate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"google.golang.org/api/storage/v1"
)

const (
	normalizedVideoBucket = "videos-normalized-3ec32eeafcfe42f28cb86296afa48673"
	compilationsBucket    = "compilations-f714ffc72eaf414ea0f51b18f4678383"
)

func init() {
	functions.HTTP("ConcatenateVideos", concatenateVideos)
}

func concatenateVideos(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	storageService, err := storage.NewService(ctx)
	if err != nil {
		log.Fatalf("Failed to create storage service: %v", err)
	}

	// Get the minimum number of videos from the environment variable, or use default value of 30
	minVideosStr := os.Getenv("MIN_VIDEOS")
	minVideos := 30
	if minVideosStr != "" {
		minVideos, err = strconv.Atoi(minVideosStr)
		if err != nil {
			log.Fatalf("Invalid MIN_VIDEOS environment variable: %v", err)
		}
	}

	// Count the number of videos in the "normalized" bucket
	objects, err := storageService.Objects.List(normalizedVideoBucket).Do()
	if err != nil {
		log.Fatalf("Failed to list objects: %v", err)
	}
	videoCount := len(objects.Items)

	if videoCount < minVideos {
		fmt.Fprintf(w, "Not enough videos to create a compilation. Found %d videos, need at least %d.", videoCount, minVideos)
		return
	}

	// Create a temporary directory to store the downloaded videos
	tempDir, err := os.MkdirTemp("", "normalized-videos")
	if err != nil {
		log.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Download the videos from the "normalized" bucket
	var videoFiles []string
	for _, object := range objects.Items {
		videoFile := filepath.Join(tempDir, object.Name)
		file, err := os.Create(videoFile)
		if err != nil {
			log.Fatalf("Failed to create file: %v", err)
		}
		defer file.Close()

		res, err := storageService.Objects.Get(normalizedVideoBucket, object.Name).Download()
		if err != nil {
			log.Fatalf("Failed to download object: %v", err)
		}
		defer res.Body.Close()

		if _, err := io.Copy(file, res.Body); err != nil {
			log.Fatalf("Failed to copy video: %v", err)
		}

		videoFiles = append(videoFiles, videoFile)
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

	outputFile := filepath.Join(tempDir, "output.mp4")

	// Run ffmpeg command to concatenate the videos together

	// Fast command but may have issues with audio/video sync
	cmd := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", videoListFile, "-c", "copy", outputFile)

	// Slow command but maintains good vid quality
	// presets: ultrafast, superfast, veryfast, faster, fast, medium, slow, slower, veryslow
	// cmd := exec.Command("ffmpeg",
	// 	"-f", "concat",
	// 	"-safe", "0",
	// 	"-i", videoListFile,
	// 	"-c:v", "libx264",
	// 	"-preset", "veryslow",
	// 	"-crf", "21",
	// 	"-pix_fmt", "yuv420p",
	// 	"-c:a", "aac",
	// 	"-ar", "48000",
	// 	"-b:a", "384k",
	// 	outputFile,
	// )

	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to run ffmpeg command: %v", err)
	}

	timestamp := time.Now().Format("20060102150405") // Format: YYYYMMDDHHmmss

	// Upload the compilation video to the "compilation" bucket with the timestamp in the filename
	outputFileData, err := os.ReadFile(outputFile)
	if err != nil {
		log.Fatalf("Failed to read output file: %v", err)
	}
	object := &storage.Object{Name: fmt.Sprintf("compilation-%s.mp4", timestamp)}
	_, err = storageService.Objects.Insert(compilationsBucket, object).Media(bytes.NewReader(outputFileData)).Do()
	if err != nil {
		log.Fatalf("Failed to upload compilation video: %v", err)
	}

	// Delete the normalized videos from the "normalized" bucket
	for _, object := range objects.Items {
		err := storageService.Objects.Delete(normalizedVideoBucket, object.Name).Do()
		if err != nil {
			log.Printf("Failed to delete object %q: %v", object.Name, err)
		}
	}

	fmt.Fprintf(w, "Compilation video created and uploaded successfully.")
}
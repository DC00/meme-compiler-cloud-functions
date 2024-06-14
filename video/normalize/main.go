package normalize

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
)

const (
	normalizedBucketName = "videos-normalized-3ec32eeafcfe42f28cb86296afa48673"
)

func init() {
	functions.CloudEvent("NormalizeVideo", normalizeVideo)
}

type StorageObjectData struct {
	Bucket string `json:"bucket,omitempty"`
	Name   string `json:"name,omitempty"`
}

func normalizeVideo(ctx context.Context, e event.Event) error {
	var data StorageObjectData
	if err := e.DataAs(&data); err != nil {
		log.Printf("Error parsing CloudEvent data: %v", err)
		return fmt.Errorf("event.DataAs: %v", err)
	}

	inputURI := fmt.Sprintf("gs://%s/%s", data.Bucket, data.Name)
	outputFilePath := fmt.Sprintf("/tmp/normalized-%s", data.Name)

	// Normalize the video using FFmpeg
	cmd := exec.Command("ffmpeg", "-i", inputURI,
		"-vf", "scale=1280:720:force_original_aspect_ratio=decrease,pad=1280:720:(ow-iw)/2:(oh-ih)/2,setsar=1,fps=30",
		"-af", "loudnorm=I=-16:TP=-1.5:LRA=11:print_format=summary,aformat=channel_layouts=stereo",
		"-c:v", "libx264", "-preset", "veryslow", "-crf", "21", "-pix_fmt", "yuv420p",
		"-c:a", "aac", "-ar", "48000", "-b:a", "384k",
		outputFilePath)

	if err := cmd.Run(); err != nil {
		log.Printf("Error running FFmpeg: %v", err)
		return fmt.Errorf("cmd.Run: %v", err)
	}

	// Upload the normalized video to the new bucket using gsutil
	gsutilCopyCmd := exec.Command("gsutil", "cp", outputFilePath, fmt.Sprintf("gs://%s/%s", normalizedBucketName, data.Name))
	if err := gsutilCopyCmd.Run(); err != nil {
		log.Printf("Error copying normalized video: %v", err)
		return fmt.Errorf("gsutilCopyCmd.Run: %v", err)
	}

	// Delete the original video file using gsutil
	gsutilDeleteCmd := exec.Command("gsutil", "rm", inputURI)
	if err := gsutilDeleteCmd.Run(); err != nil {
		log.Printf("Error deleting original video: %v", err)
		return fmt.Errorf("gsutilDeleteCmd.Run: %v", err)
	}

	// Clean up temporary output file
	os.Remove(outputFilePath)

	log.Printf("Video normalized and uploaded to bucket: %s", normalizedBucketName)
	return nil
}

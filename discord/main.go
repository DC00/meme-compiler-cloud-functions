package function

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/DC00/meme-compiler/client"
	"github.com/bwmarrin/discordgo"
)

var (
	identityToken = os.Getenv("IDENTITY_TOKEN")
)

// HandleRequest is the entry point for the Cloud Function
func HandleRequest(w http.ResponseWriter, r *http.Request) {
	var interaction discordgo.Interaction
	if err := json.NewDecoder(r.Body).Decode(&interaction); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	response := handleInteraction(interaction)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleInteraction processes the interaction from Discord
func handleInteraction(interaction discordgo.Interaction) *discordgo.InteractionResponse {
	if interaction.Type == discordgo.InteractionApplicationCommand {
		data := interaction.ApplicationCommandData()
		switch data.Name {
		case "addvideo":
			return handleAddVideo(data)
		}
	}
	return nil
}

// handleAddVideo processes the add video command
func handleAddVideo(data discordgo.ApplicationCommandInteractionData) *discordgo.InteractionResponse {
	var videoURL string
	for _, option := range data.Options {
		if option.Name == "url" {
			videoURL = option.StringValue()
			break
		}
	}

	if videoURL == "" {
		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "No video URL provided.",
			},
		}
	}

	c := client.NewClient(identityToken)

	ctx := context.Background()
	addResp, err := c.Videos.Add(ctx, &client.AddVideoRequest{
		URL: videoURL,
	})
	if err != nil {
		log.Printf("Error adding video: %v", err)
		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Error adding video: %v", err),
			},
		}
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf(addResp.Message),
		},
	}
}

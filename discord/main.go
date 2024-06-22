package discord

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/DC00/meme-compiler/client"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/bwmarrin/discordgo"
)

var botToken = os.Getenv("DISCORD_BOT_TOKEN")

func init() {
	functions.HTTP("HandleRequest", handleRequest)
}

// handleRequest is the entry point for the Cloud Function
func handleRequest(w http.ResponseWriter, r *http.Request) {
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
	// Example: Respond to a ping command
	if interaction.Type == discordgo.InteractionApplicationCommand {
		switch interaction.Data.Name {
		case "ping":
			return &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionApplicationCommandResponseData{
					Content: "Pong!",
				},
			}
		case "addvideo":
			return handleAddVideo(interaction)
		}
	}
	return nil
}

// handleAddVideo processes the add video command
func handleAddVideo(interaction discordgo.Interaction) *discordgo.InteractionResponse {
	// Assume the video URL is passed as an option
	var videoURL string
	for _, option := range interaction.Data.Options {
		if option.Name == "url" && option.Type == discordgo.ApplicationCommandOptionString {
			videoURL = option.StringValue()
			break
		}
	}

	if videoURL == "" {
		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionApplicationCommandResponseData{
				Content: "No video URL provided.",
			},
		}
	}

	c := client.NewClient("identityToken") // Replace with appropriate token management

	ctx := context.Background()
	addResp, err := c.Videos.Add(ctx, &client.AddVideoRequest{
		URL: videoURL,
	})
	if err != nil {
		log.Printf("Error adding video: %v", err)
		return &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionApplicationCommandResponseData{
				Content: fmt.Sprintf("Error adding video: %v", err),
			},
		}
	}

	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionApplicationCommandResponseData{
			Content: fmt.Sprintf("Add video response: %s", addResp.Message),
		},
	}
}

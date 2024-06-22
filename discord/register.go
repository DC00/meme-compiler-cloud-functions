// Register a slash command with the Discord API
// /addVideo

package discord

import (
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

func main() {
	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	guildID := os.Getenv("DISCORD_GUILD_ID")

	dg, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	// Define the addVideo command
	addVideoCommand := &discordgo.ApplicationCommand{
		Name:        "addVideo",
		Description: "Add a video to the meme compiler",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "url",
				Description: "A url to a funny video",
				Type:        discordgo.ApplicationCommandOptionString,
				Required:    true,
			},
		},
	}

	// Register the command
	_, err = dg.ApplicationCommandCreate(dg.State.User.ID, guildID, addVideoCommand)
	if err != nil {
		log.Fatalf("Error creating command: %v", err)
	}

	log.Println("Slash commands registered successfully.")
}

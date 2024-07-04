package main

import (
    "log"
    "os"

    "github.com/bwmarrin/discordgo"
)

func main() {
    botToken := os.Getenv("DISCORD_BOT_TOKEN")
    if botToken == "" {
        log.Fatal("No bot token provided. Set the DISCORD_BOT_TOKEN environment variable and try again.")
    }

    guildID := os.Getenv("DISCORD_GUILD_ID")
    if guildID == "" {
        log.Fatal("No guild ID provided. Set the DISCORD_GUILD_ID environment variable and try again.")
    }

    dg, err := discordgo.New("Bot " + botToken)
    if err != nil {
        log.Fatalf("Error creating Discord session: %v", err)
    }

    // Open a websocket connection to Discord and begin listening.
    err = dg.Open()
    if err != nil {
        log.Fatalf("Error opening connection: %v", err)
    }
    defer dg.Close()

    // Define the ping command
    pingCommand := &discordgo.ApplicationCommand{
        Name:        "ping",
        Description: "Ping the bot to check if it's online",
    }

    // Register the ping command
    registeredCmd, err := dg.ApplicationCommandCreate(dg.State.User.ID, guildID, pingCommand)
    if err != nil {
        log.Fatalf("Error creating ping command: %v", err)
    }

    log.Printf("Slash command registered successfully: %v", registeredCmd)

    // Define the addvideo command
    addVideoCommand := &discordgo.ApplicationCommand{
        Name:        "addvideo",
        Description: "Add a video to the meme compiler",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Name:        "url",
                Description: "A URL to a funny video",
                Type:        discordgo.ApplicationCommandOptionString,
                Required:    true,
            },
        },
    }

    // Register the addvideo command
    registeredCmd, err = dg.ApplicationCommandCreate(dg.State.User.ID, guildID, addVideoCommand)
    if err != nil {
        log.Fatalf("Error creating addvideo command: %v", err)
    }

    log.Printf("Slash command registered successfully: %v", registeredCmd)

    // Define the createcompilation command
    createCompilationCommand := &discordgo.ApplicationCommand{
        Name:        "createcompilation",
        Description: "Create a meme compilation",
    }

    // Register the createcompilation command
    registeredCmd, err = dg.ApplicationCommandCreate(dg.State.User.ID, guildID, createCompilationCommand)
    if err != nil {
        log.Fatalf("Error creating createcompilation command: %v", err)
    }

    log.Printf("Slash command registered successfully: %v", registeredCmd)
}

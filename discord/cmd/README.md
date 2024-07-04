# Discord Slash Command Registration

`register.go` registers global slash commands with Discord.

```
/ping: Pong
/addvideo [url]: Add a video to the meme compiler
/createcompilation: Creates a meme compilation
```

#### Setup
```
export DISCORD_BOT_TOKEN=myToken
go run register.go
```

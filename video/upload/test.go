package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"

    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/option"
    "google.golang.org/api/youtube/v3"
)

const (
    credentialsFile = "client-secret.json"
    tokenFile       = "token.json"
)

var scopes = []string{youtube.YoutubeUploadScope}

func main() {
    ctx := context.Background()

    // Load client secrets from credentials file
    b, err := os.ReadFile(credentialsFile)
    if err != nil {
        log.Fatalf("Unable to read client secret file: %v", err)
    }

    // Parse client secrets
    config, err := google.ConfigFromJSON(b, scopes...)
    if err != nil {
        log.Fatalf("Unable to parse client secret file to config: %v", err)
    }

    // Load token from file or prompt user to authenticate
    tok, err := tokenFromFile(tokenFile)
    if err != nil {
        tok = getTokenFromWeb(config)
        saveToken(tokenFile, tok)
    }

    // Create YouTube service
    service, err := youtube.NewService(ctx, option.WithTokenSource(config.TokenSource(ctx, tok)))
    if err != nil {
        log.Fatalf("Unable to create YouTube service: %v", err)
    }

    // Prepare video metadata
    video := &youtube.Video{
        Snippet: &youtube.VideoSnippet{
            Title:       "Test",
            Description: "test",
            Tags:        []string{"test"},
        },
        Status: &youtube.VideoStatus{
            PrivacyStatus: "private",
        },
    }

    // Open video file
    videoFile, err := os.Open("compilation.mp4")
    if err != nil {
        log.Fatalf("Unable to open video file: %v", err)
    }
    defer videoFile.Close()

    // Upload video
    response, err := service.Videos.Insert([]string{"snippet", "status"}, video).Media(videoFile).Do()
    if err != nil {
        log.Fatalf("Unable to upload video: %v", err)
    }

    fmt.Printf("Video ID: %v\n", response.Id)
}

// tokenFromFile retrieves the token from the token file
func tokenFromFile(file string) (*oauth2.Token, error) {
    f, err := os.Open(file)
    if err != nil {
        return nil, err
    }
    defer f.Close()
    tok := &oauth2.Token{}
    err = json.NewDecoder(f).Decode(tok)
    return tok, err
}

// getTokenFromWeb requests a token from the web, then returns the retrieved token
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
    authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
    fmt.Printf("Go to the following link in your browser then type the "+
        "authorization code: \n%v\n", authURL)

    var authCode string
    if _, err := fmt.Scan(&authCode); err != nil {
        log.Fatalf("Unable to read authorization code: %v", err)
    }

    tok, err := config.Exchange(context.TODO(), authCode)
    if err != nil {
        log.Fatalf("Unable to retrieve token from web: %v", err)
    }
    return tok
}

// saveToken saves a token to a file path
func saveToken(path string, token *oauth2.Token) {
    fmt.Printf("Saving credential file to: %s\n", path)
    f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
    if err != nil {
        log.Fatalf("Unable to cache oauth token: %v", err)
    }
    defer f.Close()
    json.NewEncoder(f).Encode(token)
}

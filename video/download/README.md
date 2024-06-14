# Download

Downloads a video using yt-dlp and uploads to a Google Cloud Storage bucket.

#### Explanation of the yt-dlp command:
`format=bv*[ext=mp4]+ba[ext=m4a]/b[ext=mp4]`: Enforce mp4 video and m4a audio, or best available mp4

`outputTemplate=%(extractor)s-%(id)s.%(ext)s`: platform-identifier.filetype, e.g. youtube-BaWjenozKc.mp4

The output directory can be appended to the outputTemplate. So `-o /tmp/youtube-BaWjenozKc.mp4` will store the video in the `/tmp` directory.


#### Cloud Storage
Google Cloud Run can access the storage buckets through Background context:
```
ctx := context.Background()
client, err := storage.NewClient(ctx)
```

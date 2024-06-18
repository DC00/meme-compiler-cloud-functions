# Upload

**Cloud Function** which uploads a video to Youtube. Using cloud functions because 1) they are cheaper than Cloud Run invocations and 2) Youtube's golang package is good. This means the Dockerfile in this folder is not used - GCP will wrap the function contained in `main.go` with it's own Dockerfile upon upload.


In auth code flow, you may need to change %2F to '/'
https://stackoverflow.com/questions/58209700/how-to-fix-the-malformed-auth-code-when-trying-to-refreshtoken-on-the-second-a



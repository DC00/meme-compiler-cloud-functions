# Normalize

**Cloud Function** which runs ffmpeg to normalize the size and audio of a video. Triggered by an Eventarc notification when a new object is written to the GCS bucket.

ffmpeg is included in the default runtime environment of [Go 1.22](https://cloud.google.com/functions/docs/reference/system-packages)

You can also pull the image itself [here](https://cloud.google.com/functions/docs/concepts/execution-environment).
```
docker pull \
    us-central1-docker.pkg.dev/serverless-runtimes/google-22-full/runtimes/go122:go122_20240609_1_22_3_RC00

docker run -u root -it --rm us-central1-docker.pkg.dev/serverless-runtimes/google-22-full/runtimes/go122:go122_20240609_1_22_3_RC00 /bin/bash
```




curl -m 70 -X POST https://us-east4-meme-compiler.cloudfunctions.net/mcf-normalize \
-H "Authorization: bearer $(gcloud auth print-identity-token)" \
-H "Content-Type: application/json" \
-H "ce-id: 1234567890" \
-H "ce-specversion: 1.0" \
-H "ce-type: google.cloud.storage.object.v1.finalized" \
-H "ce-time: 2020-08-08T00:11:44.895529672Z" \
-H "ce-source: //storage.googleapis.com/projects/_/buckets/videos-quarantine-2486aa1dcdb442fda0c2f090761b4479" \
-d '{
  "name": "folder/Test.cs",
  "bucket": "videos-quarantine-2486aa1dcdb442fda0c2f090761b4479",
  "contentType": "application/json",
  "metageneration": "1",
  "timeCreated": "2024-06-13T07:38:57.230Z",
  "updated": "2024-06-13T07:38:57.230Z"
}'
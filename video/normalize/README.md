# Normalize

**Cloud Function** which runs ffmpeg to normalize the size and audio of a video. Triggered by an Eventarc notification when a new object is written to the GCS bucket.

ffmpeg is included in the default runtime environment of [Go 1.22](https://cloud.google.com/functions/docs/reference/system-packages)

You can also pull the image itself [here](https://cloud.google.com/functions/docs/concepts/execution-environment).
```
docker pull \
    us-central1-docker.pkg.dev/serverless-runtimes/google-22-full/runtimes/go122:go122_20240609_1_22_3_RC00

docker run -u root -it --rm us-central1-docker.pkg.dev/serverless-runtimes/google-22-full/runtimes/go122:go122_20240609_1_22_3_RC00 /bin/bash
```

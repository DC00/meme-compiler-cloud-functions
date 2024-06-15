# Stitch

Combines normalized videos together for a final meme compilation. The Meme Compiler.

Currently using the cloud storage package to download/upload/remove videos from buckets because the gsutil sdk is not included in Cloud Functions. Cloud Functions are cheaper than Cloud Run invocations so went trying to use them as much as possible.
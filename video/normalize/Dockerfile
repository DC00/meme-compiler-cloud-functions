FROM --platform=linux/amd64 golang:1.22-bookworm as build

# Create and change to the app directory.
WORKDIR /app

# Root user to install dependencies and build code
USER root

# Expecting to copy go.mod and if present go.sum
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY main.go main.go

RUN go build -v -o server

# Use the official Debian slim image for a lean production container.
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM --platform=linux/amd64 debian:bookworm-slim as final

RUN apt-get update -qq && \
    apt-get install -y --no-install-recommends ffmpeg=7:5.1.4-0+deb12u1

# Copy the binary to production image from build stage
COPY --from=build /app/server /server

CMD ["/server"]

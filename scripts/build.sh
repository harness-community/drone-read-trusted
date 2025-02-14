#!/bin/sh
export CGO_ENABLED=0

set -e
set -x

# Build for Linux amd64
GOOS=linux GOARCH=amd64 go build -o release/linux/amd64/drone-read-trusted

# Build for Linux arm64
GOOS=linux GOARCH=arm64 go build -o release/linux/arm64/drone-read-trusted

# Build for Windows amd64
GOOS=windows GOARCH=amd64 go build -o release/windows/amd64/drone-read-trusted.exe

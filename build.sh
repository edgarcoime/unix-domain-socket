#!/bin/bash

# Check if go is installed
check_go_installed() {
  if ! command -v go &> /dev/null; then
    echo "GO is not installed in your system."
    echo "Please install GO from: https://golang.org/dl/"
    exit 1
  fi
}

check_go_installed

# Create the binaries folder if it doesn't exist
mkdir -p binaries

# compile server
go build -o binaries/ ./cmd/server/

# compile client
go build -o binaries/ ./cmd/client/

echo "Compilation complete. Client and Server programs are in the binaries/ folder."

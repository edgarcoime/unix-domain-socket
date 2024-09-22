#!/bin/bash

# Check if go is installed
check_go_installed() {
  if ! command -v go &> /dev/null; then
    echo "GO is not installed in your system."
    echo "Please install GO from: https://go.dev/dl/"
    exit 1
  fi
}

check_go_installed

# Create the binaries folder if it doesn't exist
mkdir -p bin

# compile server
go build -o bin/ ./cmd/server/

# compile client
go build -o bin/ ./cmd/client/

echo "Compilation complete. Client and Server programs are in the binaries/ folder."

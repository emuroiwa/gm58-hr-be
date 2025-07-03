#!/bin/bash

echo "Building GM58-HR Backend..."

# Clean previous builds
rm -f main

# Build the application
go build -o main cmd/server/main.go

if [ $? -eq 0 ]; then
    echo "Build successful!"
    echo "Run './main' to start the server"
else
    echo "Build failed!"
    exit 1
fi

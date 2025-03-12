#!/bin/bash

# Load environment variables
source ./env.sh

# Build and run the application
echo "Building and running the forum application..."
go build -o forum
./forum 
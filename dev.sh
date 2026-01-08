#!/bin/bash

# 1. Generate Swagger docs
echo "🚀 Generating Swagger documentation..."
swag init

# Check if swag init was successful
if [ $? -eq 0 ]; then
    echo "✅ Swagger docs generated successfully."
    
    # 2. Run the Go application
    echo "🏃 Starting the server..."
    go run main.go
else
    echo "❌ Swagger generation failed. Server will not start."
    exit 1
fi
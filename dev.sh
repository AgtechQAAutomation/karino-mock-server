#!/bin/bash

# 1. Generate Swagger docs
echo "🚀 Generating Swagger documentation..."
swag init

# Check if swag init was successful
if [ $? -eq 0 ]; then
    echo "✅ Swagger docs generated successfully."

    # 2. Run query generator ONLY if 'query' folder does NOT exist
    if [ ! -d "query" ]; then
        echo "📂 'Query' folder not found. Generating query docs..."
        go run cmd/generate/main.go
    else
        echo "📂 'Query' folder already exists. Skipping query generation."
    fi

    # 3. Run the Go application
    echo "🏃 Starting the server..."
    go run main.go
else
    echo "❌ Swagger generation failed. Server will not start."
    exit 1
fi

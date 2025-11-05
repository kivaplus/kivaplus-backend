#!/bin/bash
set -e

echo "🏗️  Building Lambda functions..."

# Create dist directory
mkdir -p dist

# Function to build and package a Lambda
build_lambda() {
    local name=$1
    local main_path=$2

    echo "📦 Building $name..."

    # Build for Linux (Lambda runtime)
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
        -ldflags="-s -w" \
        -o dist/$name \
        $main_path

    # Create deployment package
    cd dist
    zip -r ${name}.zip $name
    rm $name
    cd ..

    echo "✅ $name built successfully"
}

# Build all Lambda functions
build_lambda "auth" "./cmd/users/main.go"
build_lambda "authorizer" "./cmd/authorizer/main.go"
build_lambda "users" "./cmd/users/main.go"
build_lambda "condominiums" "./cmd/condominiums/main.go"

echo "🎉 All Lambda functions built successfully!"
ls -la dist/

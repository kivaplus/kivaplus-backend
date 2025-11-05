# Lambda Builder Dockerfile
# Builds Go Lambda functions and creates deployment packages

FROM golang:1.23 AS builder

# Install required packages
RUN apt-get update && apt-get install -y \
    git \
    zip \
    make \
    bash \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /workspace

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Create dist directory
RUN mkdir -p dist

# Build script that creates all Lambda deployment packages
RUN cat > build-lambdas.sh << 'EOF'
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
build_lambda "auth" "./cmd/auth/main.go"
build_lambda "users" "./cmd/users/main.go"
build_lambda "condominiums" "./cmd/condominiums/main.go"

echo "🎉 All Lambda functions built successfully!"
ls -la dist/
EOF

# Make build script executable
RUN chmod +x build-lambdas.sh

# Default command
CMD ["./build-lambdas.sh"]

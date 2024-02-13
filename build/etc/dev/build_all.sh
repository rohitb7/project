#!/usr/bin/env bash

if [ -z "$WORKSPACE_DIR" ]; then
    echo "WORKSPACE_DIR is not set. Please set it before running this script."
    exit 1
fi

echo "Pruning Docker system..."
docker system prune -a --force

echo "Current workspace directory: $WORKSPACE_DIR"

echo "Deleting all vendor directories from the root and submodules..."
# Find and delete vendor directories recursively from the current directory
if ! find . -type d -name "vendor" -exec rm -rf {} +; then
    echo "Failed to delete vendor directories."
    exit 1
fi
echo "All vendor directories have been deleted."

echo "Cleaning all the go cache"
if ! go clean -cache -testcache -modcache; then
    echo "Failed to clean go cache."
fi

# Define other setup paths and names
PROJECT_SUBDIR="patient-service"
PROJECT_DIR="$WORKSPACE_DIR/$PROJECT_SUBDIR"
OUTPUT_DIR="$WORKSPACE_DIR/bin"
BINARY_NAME="patient_service"

# Ensure the script does not proceed if any command fails
set -e

# Change to the workspace directory
cd "$WORKSPACE_DIR"

# Build Proto
echo "Building proto..."
BUILD_PROTO_SCRIPT_PATH="$WORKSPACE_DIR/build/etc/build_proto.sh"
chmod +x "$BUILD_PROTO_SCRIPT_PATH"
"$BUILD_PROTO_SCRIPT_PATH"

# Setup MinIO
echo "Setting up MinIO..."
MINIO_SETUP_SCRIPT_PATH="$WORKSPACE_DIR/build/etc/dev/minio_setup.sh"
chmod +x "$MINIO_SETUP_SCRIPT_PATH"
"$MINIO_SETUP_SCRIPT_PATH"

sleep 3

# Setup PostgreSQL
echo "Setting up PostgreSQL..."
POSTGRES_SETUP_SCRIPT_PATH="$WORKSPACE_DIR/build/etc/dev/postgres_setup.sh"
chmod +x "$POSTGRES_SETUP_SCRIPT_PATH"
"$POSTGRES_SETUP_SCRIPT_PATH"

sleep 3

# Setup Prometheus
#echo "Setting up Prometheus..."
#PROMETHEUS_SETUP_SCRIPT_PATH="$WORKSPACE_DIR/build/etc/prometheus_setup.sh"
#chmod +x "$PROMETHEUS_SETUP_SCRIPT_PATH"
#"$PROMETHEUS_SETUP_SCRIPT_PATH"

echo "Waiting for all services to initialize..."
sleep 3

echo "Tidying and vendoring Go modules..."

# Function to tidy and vendor for a given Go module, now exits if go mod tidy or go mod vendor fails
tidy_and_vendor() {
    local mod_dir="$1"
    echo "Processing module in $mod_dir"
    cd "$mod_dir"
    go mod tidy && go mod vendor || { echo "Failed to tidy and vendor in $mod_dir"; exit 1; }
    echo "Tidy and vendor in $mod_dir completed."
}

export -f tidy_and_vendor

# Find all go.mod files and run tidy and vendor for each, exits if find command fails
find "$WORKSPACE_DIR" -type f -name 'go.mod' -exec bash -c 'tidy_and_vendor "$(dirname {})"' \; || exit 1

echo "All Go modules have been tidied and vendored."

# Change into the project subdirectory
echo "Changing into project directory: $PROJECT_DIR"
cd "$PROJECT_DIR"

# Clean previous builds
go clean

OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Linux*)     os="linux";;
    Darwin*)    os="darwin";;
    *)          echo "Unsupported operating system: $OS"; exit 1;;
esac

case "$ARCH" in
    x86_64* | amd64*)   arch="amd64";;
    i386* | i686*)      arch="386";;
    arm64* | aarch64*)  arch="arm64";;
    *)                  echo "Unsupported architecture: $ARCH"; exit 1;;
esac

echo "Building Go application for $os/$arch..."
if ! env GO111MODULE=on GOOS=$os GOSUMDB=off GOARCH=$arch go build -o "$OUTPUT_DIR/${BINARY_NAME}"; then
    echo "Failed to build the Go application."
    exit 1
fi

echo "Binary will be in $OUTPUT_DIR/${BINARY_NAME}"
chmod +x "$OUTPUT_DIR/${BINARY_NAME}"

#..upload images to minio using the test
cd "$WORKSPACE_DIR/blob-service/s3-manager"
if ! go test -run TestUploadAllFiles; then
    echo "Failed to run TestUploadAllFiles."
    exit 1
fi

cd "$WORKSPACE_DIR"

echo "Setup complete."

#!/bin/bash

# Set environment variables
export MINIO_DOCKER_NAME=minio_server
export MINIO_VOLUME=$WORKSPACE_DIR/blob-service/minio-volume

# Stop and remove the MinIO container
echo "Stopping and removing MinIO container..."
docker stop $MINIO_DOCKER_NAME >/dev/null 2>&1
docker rm $MINIO_DOCKER_NAME >/dev/null 2>&1

# Delete the MinIO image
echo "Deleting MinIO image..."
docker rmi minio/minio >/dev/null 2>&1

# Cleanup local MinIO volume directory
echo "Cleaning up local MinIO volume directory..."
rm -rf $MINIO_VOLUME

echo "Cleanup complete"

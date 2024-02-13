#!/bin/bash

# Set environment variables
MINIO_DOCKER_NAME=minio_server
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=mybucket
MINIO_PORT=9000
MINIO_CONSOLE_PORT=9001

# Label for the Docker container
MINIO_LABEL_KEY="environment"
MINIO_LABEL_VALUE="test"

# Stop and remove any existing MinIO container
echo "Stopping and removing existing MinIO container..."
docker stop $MINIO_DOCKER_NAME > /dev/null 2>&1
docker rm $MINIO_DOCKER_NAME > /dev/null 2>&1

# Start MinIO Docker container with port forwarding and a label
echo "Starting MinIO Docker container..."
docker run -d --restart unless-stopped \
  --name $MINIO_DOCKER_NAME \
  -e "MINIO_ROOT_USER=$MINIO_ACCESS_KEY" \
  -e "MINIO_ROOT_PASSWORD=$MINIO_SECRET_KEY" \
  -p $MINIO_PORT:$MINIO_PORT \
  -p $MINIO_CONSOLE_PORT:$MINIO_CONSOLE_PORT \
  --memory 5g \
  --label $MINIO_LABEL_KEY=$MINIO_LABEL_VALUE \
  minio/minio server /data --console-address ":$MINIO_CONSOLE_PORT"

# Wait for MinIO to start up
echo "Waiting for MinIO to start..."
sleep 10

# Create the MinIO bucket using the mc command
echo "Creating bucket..."
docker exec $MINIO_DOCKER_NAME /bin/sh -c "\
  /usr/bin/mc alias set myminio http://localhost:$MINIO_PORT $MINIO_ACCESS_KEY $MINIO_SECRET_KEY && \
  /usr/bin/mc mb myminio/$MINIO_BUCKET"

# Validate the bucket creation
if [ $? -ne 0 ]; then
  echo "Failed to create bucket $MINIO_BUCKET."
#  exit 1
else
  echo "Bucket $MINIO_BUCKET created successfully."
fi


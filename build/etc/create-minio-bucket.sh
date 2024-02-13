#!/bin/bash

# Wait for MinIO to be ready
echo "Waiting for MinIO to start..."
while ! nc -z localhost 9000; do
  sleep 1
done
echo "MinIO started."

# Environment variables - replace these with actual values or pass them to the script
MINIO_DOCKER_NAME=minio_server # Docker container name
MINIO_PORT=9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=mybucket

# Create the MinIO bucket
echo "Creating bucket $MINIO_BUCKET..."
docker exec $MINIO_DOCKER_NAME /bin/sh -c "\
  /usr/bin/mc alias set myminio http://localhost:$MINIO_PORT $MINIO_ACCESS_KEY $MINIO_SECRET_KEY && \
  /usr/bin/mc mb myminio/$MINIO_BUCKET"

# Validate the bucket creation
if [ $? -ne 0 ]; then
  echo "Failed to create bucket $MINIO_BUCKET."
  exit 1
else
  echo "Bucket $MINIO_BUCKET created successfully."
fi

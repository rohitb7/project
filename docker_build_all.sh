#!/bin/bash

# Exit the script on any command failure
set -e

# Step 1: Clean up Docker environment (WARNING: This is a destructive operation)
echo "Pruning Docker system..."
docker system prune -a --force

# Step 2: Build the Docker image without using cache
echo "Building Docker image..."
docker build --no-cache -t patient-image:latest .

# Step 3: Start up services defined in Docker Compose file
echo "Starting up services..."
docker-compose up -d

# Give services time to start
echo "Waiting for services to start...1 min "
sleep 60  # Adjust sleep duration as necessary

# Step 4: Initialize PostgreSQL database
echo "Initializing PostgreSQL database..."
postgres_container=$(docker ps --filter "name=project-postgres-patients-1" --format "{{.Names}}" | head -n 1 || true)

if [ -z "$postgres_container" ]; then
    echo "Postgres container not found. Exiting."
    exit 1
fi

docker exec -i $postgres_container psql -U postgres -d patients_db < ./build/etc/postgres_schema.sql || {
    echo "Failed to initialize PostgreSQL database. Exiting."
    exit 1
}

# Step 5: Copy files to MinIO and upload them to a bucket
echo "Handling MinIO uploads..."

# Find the MinIO container
minio_container=$(docker ps --filter "ancestor=minio/minio" --format "{{.Names}}" | head -n 1 || true)

if [ -z "$minio_container" ]; then
    echo "MinIO container not found. Exiting."
    exit 1
fi

# Define local directory and bucket name
local_directory="./temperory-list-of-images"  # Adjust based on actual directory path
bucket_name="mybucket"
temp_container_dir="/var/tmp/uploads"

# Create temporary directory in MinIO container
docker exec $minio_container mkdir -p $temp_container_dir

# Copy files from local directory to the temporary directory in MinIO container
docker cp $local_directory/. $minio_container:$temp_container_dir

# Configure mc with the MinIO server (replace YOUR_ACCESS_KEY and YOUR_SECRET_KEY with actual values)
#docker exec $minio_container /bin/sh -c "\
#    mc alias set myminio http://localhost:9000 minioadmin minioadmin && \
#    mc mb myminio/$bucket_name --ignore-existing && \
#    mc cp $temp_container_dir/* myminio/$bucket_name/"

# Copy policy.json file to the MinIO container
docker cp ./policy.json $minio_container:/var/log/policy.json

docker exec $minio_container /bin/sh -c "\
    mc alias set myminio http://localhost:9000 minioadmin minioadmin && \
    mc admin policy create myminio full-access /var/log/policy.json && \
    mc admin user add myminio minioadmin minioadmin && \
    mc admin policy attach myminio full-access user=minioadmin && \
    mc mb myminio/$bucket_name --ignore-existing && \
    mc policy set public myminio/$bucket_name && \
    mc cp $temp_container_dir/* myminio/$bucket_name/"

# Note: The `--ignore-existing` flag for `mc mb` command avoids error if the bucket already exists.

# Optional: Cleanup temporary directory in MinIO container
#docker exec $minio_container /bin/sh -c "rm -rf $temp_container_dir"

echo "Script completed successfully."

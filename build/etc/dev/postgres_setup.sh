#!/bin/bash

# Configuration variables
CONTAINER_NAME="postgres-patients-debian"
DB_NAME="patients_db"
DB_USER="postgres"
DB_PASSWORD="mysecretpassword"
SCHEMA_FILE_PATH="$WORKSPACE_DIR/build/etc/postgres_schema.sql"
HOST_PORT=5432  # Port on the host machine
CONTAINER_PORT=5432  # Port inside the container
LABEL_KEY="environment"
LABEL_VALUE="test"

# Pull the latest PostgreSQL image
echo "Pulling the latest PostgreSQL Docker image..."
docker pull postgres

# Check for existing container and stop it
if [ $(docker ps -a -q -f name=^/${CONTAINER_NAME}$) ]; then
    echo "Stopping and removing existing Docker container with name ${CONTAINER_NAME}..."
    docker stop ${CONTAINER_NAME}
    docker rm ${CONTAINER_NAME}
fi

# Run PostgreSQL Docker container with port forwarding and a label
echo "Running new Docker container named ${CONTAINER_NAME}..."
docker run --name ${CONTAINER_NAME} \
  -e POSTGRES_PASSWORD=${DB_PASSWORD} \
  -d -p ${HOST_PORT}:${CONTAINER_PORT} \
  --label ${LABEL_KEY}=${LABEL_VALUE} \
  postgres


# Wait for the container to start
echo "Waiting for PostgreSQL to initialize..."
sleep 15

# Create the database
echo "Creating database ${DB_NAME}..."
docker exec -it ${CONTAINER_NAME} psql -U ${DB_USER} -c "CREATE DATABASE ${DB_NAME};"

# Copy schema file to container and create schema and tables
echo "Creating schema and tables from ${SCHEMA_FILE_PATH}..."
docker cp ${SCHEMA_FILE_PATH} ${CONTAINER_NAME}:/${DB_NAME}_schema.sql
docker exec -it ${CONTAINER_NAME} psql -U ${DB_USER} -d ${DB_NAME} -f /${DB_NAME}_schema.sql

echo "PostgreSQL has been successfully set up with a new database and schema."

#!/bin/bash

# Configuration variables
CONTAINER_NAME="postgres-patients"
DB_NAME="patients_db"
DB_USER="postgres"
DB_PASSWORD="mysecretpassword"
SCHEMA_FILE_PATH="./postgres_schema.sql"
HOST_PORT=5432  # Port on the host machine
CONTAINER_PORT=5432  # Port inside the container

# Check if the container exists
if [ $(docker ps -a -q -f name=^/${CONTAINER_NAME}$) ]; then
    # Stop the container
    echo "Stopping the Docker container ${CONTAINER_NAME}..."
    docker stop ${CONTAINER_NAME}

    # Remove the container
    echo "Removing the Docker container ${CONTAINER_NAME}..."
    docker rm ${CONTAINER_NAME}
else
    echo "The Docker container ${CONTAINER_NAME} does not exist."
fi

#!/bin/bash

# Set environment variables
PROMETHEUS_DOCKER_NAME=prometheus_server
PROMETHEUS_CONFIG_FILE=./prometheus.yml
PROMETHEUS_PORT=9090
PROMETHEUS_DATA_DIR=./prometheus_data

# Label for the Docker container
PROMETHEUS_LABEL_KEY="environment"
PROMETHEUS_LABEL_VALUE="test"

# Stop and remove any existing Prometheus container
echo "Stopping and removing existing Prometheus container..."
docker stop $PROMETHEUS_DOCKER_NAME > /dev/null 2>&1
docker rm $PROMETHEUS_DOCKER_NAME > /dev/null 2>&1

# Create a Prometheus data directory and configuration file if not exists
mkdir -p $PROMETHEUS_DATA_DIR

if [ ! -f $PROMETHEUS_CONFIG_FILE ]; then
cat << EOF > $PROMETHEUS_CONFIG_FILE
global:
  scrape_interval:     15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
EOF
fi

# Start Prometheus Docker container with port forwarding and a label
echo "Starting Prometheus Docker container..."
docker run -d --restart unless-stopped \
  --name $PROMETHEUS_DOCKER_NAME \
  -p $PROMETHEUS_PORT:9090 \
  -v $(pwd)/$PROMETHEUS_CONFIG_FILE:/etc/prometheus/prometheus.yml \
  -v $(pwd)/$PROMETHEUS_DATA_DIR:/prometheus \
  --label $PROMETHEUS_LABEL_KEY=$PROMETHEUS_LABEL_VALUE \
  prom/prometheus

# Validate Prometheus startup
if [ $? -ne 0 ]; then
  echo "Failed to start Prometheus."
  exit 1
else
  echo "Prometheus started successfully."
fi

echo "Prometheus is running on http://localhost:$PROMETHEUS_PORT"

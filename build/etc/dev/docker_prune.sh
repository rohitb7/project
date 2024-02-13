#!/bin/bash

LABEL_KEY="environment"
LABEL_VALUE="test"

# Prune containers with the specified label
echo "Pruning containers with label ${LABEL_KEY}=${LABEL_VALUE}..."
docker container prune --force --filter "label=${LABEL_KEY}=${LABEL_VALUE}"

# Prune images with the specified label
echo "Pruning images with label ${LABEL_KEY}=${LABEL_VALUE}..."
docker image prune --all --force --filter "label=${LABEL_KEY}=${LABEL_VALUE}"


#!/bin/bash

# Define the base directory containing your Kubernetes YAML files
BASE_DIRECTORY="./kubernetes"

# Specify the namespace
NAMESPACE="my-namespace"

# Start Minikube if it's not already running
minikube start

# Check if the specified namespace exists, create if it doesn't
if ! kubectl get namespace "$NAMESPACE" > /dev/null 2>&1; then
    echo "Namespace '$NAMESPACE' does not exist. Creating..."
    kubectl create namespace "$NAMESPACE"
fi

# Set kubectl to use the specified namespace by default in the current context
kubectl config set-context --current --namespace="$NAMESPACE"

# Check if the base directory exists
if [ ! -d "$BASE_DIRECTORY" ]; then
  echo "Directory '$BASE_DIRECTORY' does not exist."
  exit 1
fi

# Apply all .yaml files found within the base directory and its subdirectories
find "$BASE_DIRECTORY" -type f -name "*.yaml" -print0 | while IFS= read -r -d '' file; do
    echo "Applying $file in namespace $NAMESPACE"
    kubectl apply -f "$file"
    if [ $? -ne 0 ]; then
        echo "Failed to apply $file"
        exit 1
    fi
done

echo "All .yaml files have been applied in namespace $NAMESPACE."

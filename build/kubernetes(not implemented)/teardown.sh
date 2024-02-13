#!/bin/bash

# Specify the namespace from which to delete all resources
NAMESPACE="my-namespace"

# Delete all resources in the specified namespace
kubectl delete all --all -n "$NAMESPACE"

# Optionally, delete the namespace itself
# Be cautious with this command as it will remove the namespace entirely
kubectl delete namespace "$NAMESPACE"

# To stop Minikube (optional, only if you're finished using Minikube and want to stop it)
#minikube stop

# To delete the Minikube cluster (optional, only if you want to remove Minikube completely)
#minikube delete

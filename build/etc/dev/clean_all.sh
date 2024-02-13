#!/usr/bin/env bash
echo "Running go clean,go vet,go fmt"
rm -rf vendor
go clean
go mod tidy
go mod vendor
go vet $(go list ./... | egrep -v vendor/)
go fmt $(go list ./... | egrep -v vendor/)
echo " cleaning all containers "
docker container  prune -f
docker container stop  $(docker container ls -aq)
docker container rm -fv $(docker container ls -aq)

echo "Clean Postgres and Minio containers"
etc/postgres_cleanup.sh
etc/minio_cleanup.sh

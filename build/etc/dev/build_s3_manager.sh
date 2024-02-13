#!/usr/bin/env bash
echo "${OUTPUT}"
cd $GOPATH/project/patient-service/storage/s3-manager
env GO111MODULE=on GOSUMDB=off go clean; env GO111MODULE=on GOOS=linux GOSUMDB=off GOARCH=amd64 go build -o s3_manager

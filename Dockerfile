# Use a specific version of the official Go image as the base
FROM golang:1.21.4-alpine

# Install Git (if not already installed)
RUN apk add --no-cache git
#
RUN apk add --no-cache bash

# Install Git and Protobuf Compiler
RUN apk add --no-cache git build-base

RUN apk add --no-cache protobuf protobuf-dev

# Set the working directory in the container
WORKDIR /app

# Copy the entire project into the container
COPY . ./project


# Install protoc-gen-go, protoc-gen-go-grpc, protoc-gen-grpc-gateway, and protoc-gen-openapiv2 plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest && \
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# Add $GOPATH/bin to PATH to ensure protoc can find the plugins
ENV PATH="$PATH:$(go env GOPATH)/bin"

# Copy the 'googleapis' directory
#COPY ./project/protos/googleapis ./project/protos/googleapis

# Copy all .proto files from the project's protos directory
#COPY ./protos/*.proto ./project/protos/

# Fetch the specified Go packages
#RUN go get go.opencensus.io@v0.23.0 && \

#    go get go.opencensus.io@v0.24.0 && \
#    go get google.golang.org/grpc && \
#    go get gopkg.in/yaml.v2 && \
#    go get dmitri.shuralyov.com/gpu/mtl && \
#    go get go.opentelemetry.io/otel && \
#    go get honnef.co/go/tools && \
#    go get github.com/golang/protobuf && \
#    go get golang.org/x/net && \
#    go get rsc.io/quote



WORKDIR /app
RUN protoc -I. \
       -I./project/protos/googleapis \
       --go_out=. \
       --go_opt=paths=source_relative \
       --go-grpc_out=require_unimplemented_servers=false:. \
       --go-grpc_opt=paths=source_relative \
       --grpc-gateway_out=. \
       --grpc-gateway_opt=paths=source_relative \
       --openapiv2_out=. \
       --openapiv2_opt=logtostderr=true \
       ./project/protos/*.proto



# Vendor dependencies for all Go modules
RUN find ./project -name "go.mod" -exec sh -c 'echo "Running go mod tidy && go mod vendor in $(dirname {})" && cd $(dirname {}) && go mod tidy && go mod vendor' \;

# Handle the patient-service
WORKDIR ./project/patient-service

# Run go mod tidy and go mod vendor for patient-service
RUN go mod tidy && go mod vendor

# Build the patient-service Go application ..wrong location..needs to change

RUN GOOS=linux GOARCH=amd64 go build -o ./project/bin/patient_service_binary .

# The final command to run your application
CMD ["./project/bin/patient_service_binary"]

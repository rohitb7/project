#!/bin/bash

cd $WORKSPACE_DIR/protos && \
protoc -I. \
       -I./googleapis \
       --go_out=. \
       --go_opt=paths=source_relative \
       --go-grpc_out=require_unimplemented_servers=false:. \
       --go-grpc_opt=paths=source_relative \
       --grpc-gateway_out=. \
       --grpc-gateway_opt=paths=source_relative \
       --openapiv2_out=. \
       --openapiv2_opt=logtostderr=true \
       ./*.proto  # This specifies

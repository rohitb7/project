module www.rvb.com/patient-service

go 1.21.4

require (
	github.com/dlintw/goconf v0.0.0-20120228082610-dcc070983490
	github.com/golang/protobuf v1.5.3
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.1
	github.com/lib/pq v1.10.9
	github.com/rs/cors v1.10.1
	github.com/sirupsen/logrus v1.9.3
	google.golang.org/grpc v1.61.0
	www.rvb.com/blob-service/core v0.0.0-00010101000000-000000000000
	www.rvb.com/blob-service/s3-manager v0.0.0-00010101000000-000000000000
	www.rvb.com/commonutils v0.0.0-00010101000000-000000000000
	www.rvb.com/protos v0.0.0-00010101000000-000000000000
)

require (
	github.com/aws/aws-sdk-go-v2 v1.24.1 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.5.4 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.26.6 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.16.16 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.14.11 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.15.15 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.2.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.5.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.7.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.2.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.10.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.2.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.10.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.16.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.48.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.18.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.21.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.26.7 // indirect
	github.com/aws/smithy-go v1.19.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/prometheus/client_golang v1.18.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto v0.0.0-20240125205218-1f4bbc51befe // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240205150955-31a09d347014 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240125205218-1f4bbc51befe // indirect
	google.golang.org/protobuf v1.32.0 // indirect
	www.rvb.com/blob-service/s3-conf v0.0.0-00010101000000-000000000000 // indirect
)

//replace www.rvb.com/patient-service => ../image-service

replace www.rvb.com/blob-service/core => ../blob-service/core

replace www.rvb.com/blob-service/s3-conf => ../blob-service/s3-conf

replace www.rvb.com/blob-service/s3-manager => ../blob-service/s3-manager

replace www.rvb.com/commonutils => ./../commonutils

replace www.rvb.com/metamonitor => ../metamonitor

replace www.rvb.com/protos => ./../protos

//metamonitor "www.rvb.com/patient-service/north-bound/metamonitor"
//protoenumutils "www.rvb.com/patient-service/protoenumutils"

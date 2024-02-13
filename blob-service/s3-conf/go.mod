module www.rvb.com/blob-service/s3-conf

go 1.21.4

require (
	github.com/aws/aws-sdk-go-v2 v1.11.1
	github.com/dlintw/goconf v0.0.0-20120228082610-dcc070983490
	github.com/sirupsen/logrus v1.9.3
	www.rvb.com/commonutils v0.0.0-00010101000000-000000000000
)

require (
	github.com/aws/smithy-go v1.9.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
)

//replace www.rvb.com/blob-service/s3-conf => ../s3-conf

replace www.rvb.com/blob-service/core => ./../core

replace www.rvb.com/commonutils => ./../../commonutils

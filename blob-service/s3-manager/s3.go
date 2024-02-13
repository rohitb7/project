package s3_manager

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	log "github.com/sirupsen/logrus"
	"reflect"
	"runtime"
)

const (
	AWSCloud = "AWSCloud"
	OnPremS3 = "OnPremS3"
)

// GetFunctionName Get a function Name given a function
func GetFunctionName(function interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name()
}

var clientMain *s3.Client
var s3ManagerMain *S3Manager

func getClient() *s3.Client {
	return clientMain
}

func getS3ManagerMain() *S3Manager {
	return s3ManagerMain
}

// NewS3Client
func NewS3Client() error {

	f := log.Fields{"S3Provider": getS3ManagerMain().S3Option.Config.S3Provider,
		"URL":    getS3ManagerMain().S3Option.Config.URL,
		"Region": getS3ManagerMain().S3Option.Config.Region,
		"Bucket": getS3ManagerMain().S3Option.Config.Bucket}
	log.Info(f, "NewS3Client started")

	var err error
	defer func() {
		if err != nil {
			log.Error(f, log.Fields{"error": err}, "creating clientMain failed")
		} else {
			log.Info(f, "S3 Client Created")
		}
	}()

	if getS3ManagerMain().S3Option.Config.Region == "" {
		log.Error(log.Fields{"error": err}, "Region can't be empty for data store")
		err = fmt.Errorf("Region can't be empty for data store %s", getS3ManagerMain().S3Option.Config.S3Provider)
		return err
	}
	if getS3ManagerMain().S3Option.Config.AccessKey == "" {
		log.Error(log.Fields{"error": err}, "AccessKey can't be empty for data store")
		err = fmt.Errorf("AccessKey can't be empty for data store %s", getS3ManagerMain().S3Option.Config.S3Provider)
		return err
	}

	if getS3ManagerMain().S3Option.Config.AccessSecret == "" {
		log.Error(log.Fields{"error": err}, "AccessSecret can't be empty for data store")
		err = fmt.Errorf("AccessSecret can't be empty for data store %s", getS3ManagerMain().S3Option.Config.S3Provider)
		return err
	}

	var epResolver aws.EndpointResolverFunc

	if getS3ManagerMain().S3Option.Config.Mode == "docker" {

		log.Info("******DOCKER******")

		// this needs to be done fo r
		epResolver = aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			log.Info("Using Non AWS S3 with local endpoint")
			localURL := "http://localhost:9000" // hardcoded for minio.Local URL for MinIO
			ep := aws.Endpoint{
				PartitionID:       "aws",
				URL:               localURL,
				SigningRegion:     getS3ManagerMain().S3Option.Config.Region,
				HostnameImmutable: true,
				Source:            aws.EndpointSourceCustom,
			}
			return ep, nil
		})
	} else {

		log.Info("******DEV******")

		epResolver = aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			// This handles AWS VPC endpoints resolution
			// Ref https://docs.aws.amazon.com/AmazonS3/latest/userguide/privatelink-interface-endpoints.html#accessing-bucket-and-aps-from-interface-endpoints
			// Ref https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/endpoints/
			if getS3ManagerMain().S3Option.Config.S3Provider != AWSCloud {
				log.Info("Using Non AWS S3 with custom endpoint")
				ep := aws.Endpoint{
					PartitionID:       "aws",
					URL:               getS3ManagerMain().S3Option.Config.URL,
					SigningRegion:     getS3ManagerMain().S3Option.Config.Region,
					HostnameImmutable: true,
					Source:            aws.EndpointSourceCustom,
				}
				return ep, nil
			} else {
				if getS3ManagerMain().S3Option.Config.URL != "" {
					log.Info("Using AWSCloud with custom endpoint")
					ep := aws.Endpoint{
						PartitionID:       "aws",
						URL:               getS3ManagerMain().S3Option.Config.URL,
						SigningRegion:     getS3ManagerMain().S3Option.Config.Region,
						HostnameImmutable: false,
					}
					return ep, nil
				} else {
					log.Info("Using AWSCloud with default endpoint, url is empty")
				}
			}
			// returning EndpointNotFoundError will allow the service to fallback to it's default resolution
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		})
	}

	credProvider := credentials.StaticCredentialsProvider{
		Value: aws.Credentials{
			AccessKeyID:     getS3ManagerMain().S3Option.Config.AccessKey,
			SecretAccessKey: getS3ManagerMain().S3Option.Config.AccessSecret,
			Source:          "static credentials",
		},
	}

	ctx, cancel := context.WithCancel(getS3ManagerMain().Ctx)
	defer cancel()

	var cfg aws.Config

	// Create AWS config
	cfg, err = config.LoadDefaultConfig(ctx,
		config.WithEndpointResolver(epResolver),
		config.WithCredentialsProvider(credProvider),
		config.WithRegion(getS3ManagerMain().S3Option.Config.Region))

	if err != nil {
		log.Info(log.Fields{"epResolver": epResolver, "credProvider": credProvider, "Region": getS3ManagerMain().S3Option.Config.Region},
			"LoadDefaultConfig Failed")
		return err
	}

	log.Info(log.Fields{"S3Provider": getS3ManagerMain().S3Option.Config.S3Provider, "Region": getS3ManagerMain().S3Option.Config.Region},
		"NewS3Client Config datastore")

	cl := s3.NewFromConfig(cfg, func(options *s3.Options) {})

	s3.NewDefaultEndpointResolver()

	// set aws config
	s3ManagerMain.S3Option.AwsCfg = &cfg
	clientMain = cl

	return nil
}

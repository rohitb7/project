package s3_manager

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	log "github.com/sirupsen/logrus"
	"www.rvb.com/commonutils"
)

// CreateBucket: COMMENT:NOT USED : HARDCODED BUCKET FOR NOW
func CreateBucket(bucketName *string) (*s3.CreateBucketOutput, error) {

	s3Man := getS3ManagerMain()
	clMan := getClient()

	f := log.Fields{
		"Bucket ":    bucketName,
		"S3Provider": s3Man.S3Option.Config.S3Provider,
		"URL":        s3Man.S3Option.Config.URL,
	}

	log.Info(f, "CreateBucket started")

	var err error

	if s3Man.S3Option.AwsCfg == nil {
		err = fmt.Errorf("aws config is nil")
		log.WithFields(log.Fields{"error": err}).Error("failed to get clientMain object")
		return nil, err
	}
	bc := types.CreateBucketConfiguration{
		LocationConstraint: types.BucketLocationConstraint(s3Man.S3Option.AwsCfg.Region),
	}
	input := &s3.CreateBucketInput{
		Bucket:                    bucketName,
		CreateBucketConfiguration: &bc,
	}

	var resp *s3.CreateBucketOutput
	var retryFn = func() error {
		ctx, cancel := context.WithTimeout(s3Man.S3Option.Ctx, s3Man.S3ManagerConfig.ContextTimeout)
		defer cancel()
		resp, err = clMan.CreateBucket(ctx, input, func(options *s3.Options) {})
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("bucket create failed")
		}
		return err
	}

	err = commonutils.Retry(s3Man.S3ManagerConfig.RetryCount, s3Man.S3ManagerConfig.RetryDelay, retryFn, IsRetryableError, commonutils.LinearBackoff)

	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{"Bucket": bucketName, "location": s3Man.S3Option.Config.Region}).Info("Created bucket")
	return resp, nil
}

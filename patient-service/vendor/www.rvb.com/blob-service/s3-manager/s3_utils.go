package s3_manager

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	log "github.com/sirupsen/logrus"
	"www.rvb.com/blob-service/core"
	"www.rvb.com/commonutils"
)

const (
	FileModTime = "file_mod_time"
)

func getFileInfoFromStorage(s3m *S3Manager, ctx context.Context, remotePathKey string) (*core.FileInfo, error) {

	var err error

	clMan := getClient()

	params := &s3.HeadObjectInput{
		Bucket: aws.String(s3m.S3Option.Config.Bucket),
		Key:    aws.String(remotePathKey),
	}

	var resp *s3.HeadObjectOutput
	var retryFn = func() error {
		retryCtx, cancel := context.WithTimeout(ctx, s3m.S3ManagerConfig.ContextTimeout)
		defer cancel()
		resp, err = clMan.HeadObject(retryCtx, params)
		return err
	}

	err = commonutils.Retry(s3m.S3ManagerConfig.RetryCount, s3m.S3ManagerConfig.RetryDelay, retryFn,
		IsRetryableError, commonutils.LinearBackoff)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("failed to get object")
		return nil, err
	}

	fileInfo := &core.FileInfo{
		Size:     *resp.ContentLength,
		MetaData: resp.Metadata,
	}

	return fileInfo, nil
}

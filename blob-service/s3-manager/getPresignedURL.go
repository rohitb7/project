package s3_manager

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	log "github.com/sirupsen/logrus"
	"time"
	"www.rvb.com/blob-service/core"
	protos "www.rvb.com/protos"
)

// getPresignedURL generates a presigned URL for uploading or downloading an object from an S3 bucket.
func getPresignedURL(blobCtx core.BlobContext, ctx context.Context, expiry time.Duration, operation protos.PresignedOperation) (string, error) {
	client := getClient()

	log.Info("getPresignedURL called")

	log.WithFields(log.Fields{
		"bucket": blobCtx.HierarchyIdentifier.Bucket,
		"key":    blobCtx.RemotePathKey,
	}).Info("Generating presigned URL")

	// Create a presign client using the existing S3 client
	presignClient := s3.NewPresignClient(client)

	var input interface{}

	switch operation {
	case protos.PresignedOperation_DOWNLOAD_OPERATION:

		input = &s3.GetObjectInput{
			Bucket: aws.String(blobCtx.HierarchyIdentifier.Bucket),
			Key:    aws.String(blobCtx.RemotePathKey),
		}
	case protos.PresignedOperation_UPLOAD_OPERATION:

		input = &s3.PutObjectInput{
			Bucket: aws.String(blobCtx.HierarchyIdentifier.Bucket),
			Key:    aws.String(blobCtx.RemotePathKey),
		}
	default:
		return "", fmt.Errorf("unknown operation: %d", operation)
	}

	// Create a presigned URL
	var presignedReq *v4.PresignedHTTPRequest
	var err error
	switch operation {
	case protos.PresignedOperation_DOWNLOAD_OPERATION:
		//Check if the file exists in the bucket

		//TODO:  there a rate limiter enforced by minio?

		//_, err := client.HeadObject(ctx, &s3.HeadObjectInput{
		//	Bucket: aws.String(blobCtx.HierarchyIdentifier.Bucket),
		//	Key:    aws.String(blobCtx.RemotePathKey),
		//})
		//if err != nil {
		//	if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NotFound" {
		//		return "", fmt.Errorf("file does not exist: %v", err)
		//	}
		//	return "", fmt.Errorf("error checking file existence: %v", err)
		//}

		//File exists, proceed to generate presigned URL

		presignedReq, err = presignClient.PresignGetObject(ctx, input.(*s3.GetObjectInput), func(o *s3.PresignOptions) {
			o.Expires = expiry
		})
	case protos.PresignedOperation_UPLOAD_OPERATION:

		log.Info("CAME HERE 5")

		presignedReq, err = presignClient.PresignPutObject(ctx, input.(*s3.PutObjectInput), func(o *s3.PresignOptions) {
			o.Expires = expiry
		})
	}

	if err != nil {
		log.WithFields(log.Fields{
			"bucket": blobCtx.HierarchyIdentifier.Bucket,
			"key":    blobCtx.RemotePathKey,
			"error":  err,
		}).Error("Failed to create presigned URL")
		return "", fmt.Errorf("failed to create presigned URL: %v", err)
	}

	log.WithFields(log.Fields{
		"bucket": blobCtx.HierarchyIdentifier.Bucket,
		"key":    blobCtx.RemotePathKey,
		"url":    presignedReq,
	}).Info("Presigned URL generated successfully")

	return presignedReq.URL, nil
}

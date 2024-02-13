package core

import (
	"context"
	"time"
	protos "www.rvb.com/protos"
)

// Common sensor manager interface functions
type StorageManager interface {
	CreateBucket(bucketName string) error

	PutBlob(sourcePath string, blobCtx BlobContext, ctx context.Context, opResponse chan StorageManagerOperationEvent,
		progressResponse chan StorageManagerProgressEvent)

	GetBlob(downloadPath string, blobCtx BlobContext, ctx context.Context,
		opResponse chan StorageManagerOperationEvent,
		progressResponse chan StorageManagerProgressEvent)

	GetPresignedURL(blobCtx BlobContext, ctx context.Context, expiry time.Duration, operation protos.PresignedOperation) (string, error)

	GetBucket() string
}

package s3_manager

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	log "github.com/sirupsen/logrus"
	"math"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"www.rvb.com/blob-service/core"
	"www.rvb.com/commonutils"
)

// PutBlob // generic upload
func PutBlob(sourcePath string, blobCtx core.BlobContext, ctx context.Context, opResponse chan core.StorageManagerOperationEvent, progressResponse chan core.StorageManagerProgressEvent) error {
	s3Man := getS3ManagerMain()
	clMan := getClient()

	bucket := blobCtx.HierarchyIdentifier.Bucket

	f := log.Fields{
		"S3Provider":    s3Man.S3Option.Config.S3Provider,
		"Endpoint":      s3Man.S3Option.Config.URL,
		"sourcePath":    sourcePath,
		"remotePathKey": blobCtx.RemotePathKey,
		"Bucket":        bucket,
		"Context":       ctx,
	}
	log.WithFields(f).Info("Single file PutBlob started")

	defer close(progressResponse)
	defer close(opResponse)

	var err error
	var info commonutils.FileMetaInfo

	//Defer is scheduled at the start to be called at the end the function returns
	defer func(startTime time.Time) {
		// Error counter
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("failed to stat file") // parameter in label values should match the array size in metric definition
			core.StorageMetaMonitor.GetAPICounter().
				WithLabelValues(commonutils.GetFunctionName(PutBlob), core.MetaMonitorFAIL).Inc()
		} else {
			core.StorageMetaMonitor.GetAPICounter().
				WithLabelValues(commonutils.GetFunctionName(PutBlob), core.MetaMonitorOK).Inc()
		}

	}(time.Now())

	info, err = commonutils.GetFileMetaInfo(sourcePath)
	log.WithFields(log.Fields{"info": info}).Info("GetFileMetaInfo")
	if err != nil {
		opResponse <- core.StorageManagerOperationEvent{
			Operation: core.OperationFailure,
			Err:       err,
		}
		log.WithFields(log.Fields{"err": err}).Error("failed to stat file")
		return err
	}

	file, err := os.Open(sourcePath)
	if err != nil {
		opResponse <- core.StorageManagerOperationEvent{
			Operation: core.OperationFailure,
			Err:       err,
		}
		log.WithFields(log.Fields{"error": err}).Error("failed to get file")
		return err
	}
	defer file.Close()

	uploader := manager.NewUploader(clMan, func(u *manager.Uploader) {})

	var allowed map[uint32]uint32
	allowed = map[uint32]uint32{0: 0, 25: 25, 50: 50, 75: 75, 100: 100}

	// no need, just for progress
	reader := &CustomReader{
		fp:               file,
		size:             info.Size,
		progressResponse: progressResponse,
		allowed:          allowed,
	}

	currentTimestamp := time.Now().Unix()

	cT := fmt.Sprintf("%d", currentTimestamp)

	var metaData = map[string]string{}

	// metadata keys has to be all smallcase
	metaData[FileModTime] = cT

	input := &s3.PutObjectInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(blobCtx.RemotePathKey),
		Body:         reader,
		Metadata:     metaData,
		StorageClass: types.StorageClass(strings.ToUpper(s3Man.S3ManagerConfig.S3StorageClass)),
	}
	var resp *manager.UploadOutput
	var retryFn = func() (err error) {
		retryCtx, cancel := context.WithTimeout(ctx, s3Man.S3ManagerConfig.MaxUploadContextTimeout)
		defer cancel()
		resp, err = uploader.Upload(retryCtx, input)
		return err
	}

	err = commonutils.Retry(s3Man.S3ManagerConfig.RetryCount, s3Man.S3ManagerConfig.RetryDelay,
		retryFn, IsRetryableError,
		commonutils.LinearBackoff)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("failed to upload file")
		opResponse <- core.StorageManagerOperationEvent{
			Operation: core.OperationFailure,
			Err:       err,
		}
		return err
	}

	f = log.Fields{
		"sourcePath": sourcePath,
		"bucket":     bucket,
		"Location":   resp.Location,
	}

	opResponse <- core.StorageManagerOperationEvent{
		Operation: core.OperationSuccess,
		Err:       nil,
	}

	log.WithFields(f).Info("File Uploaded successfully")

	return nil
}

type CustomReader struct {
	fp               *os.File
	size             uint64
	read             uint64
	progressResponse chan core.StorageManagerProgressEvent
	allowed          map[uint32]uint32
	lock             sync.RWMutex
}

func (r *CustomReader) Read(p []byte) (int, error) {
	return r.fp.Read(p)
}

func (r *CustomReader) ReadAt(p []byte, off int64) (int, error) {
	n, err := r.fp.ReadAt(p, off)
	if err != nil {
		return n, err
	}

	atomic.AddUint64(&r.read, uint64(n))

	pD := float64(int(float32(r.read*100/2) / float32(r.size))) // 1.23

	if math.Mod(pD, 25) == 0 {
		r.lock.Lock()
		_, exist := r.allowed[uint32(pD)]
		if exist {
			r.progressResponse <- core.StorageManagerProgressEvent{
				Progress: pD,
			}
			delete(r.allowed, uint32(pD))
		}
		r.lock.Unlock()
	}

	return n, err
}

func (r *CustomReader) Seek(offset int64, whence int) (int64, error) {
	return r.fp.Seek(offset, whence)
}

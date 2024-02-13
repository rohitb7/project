package s3_manager

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"www.rvb.com/blob-service/core"
	"www.rvb.com/commonutils"
)

type progressWriter struct {
	written          int64
	writer           io.WriterAt
	size             int64
	progressResponse chan core.StorageManagerProgressEvent
	allowed          map[uint32]uint32
	lock             sync.RWMutex
}

func (pw *progressWriter) WriteAt(p []byte, off int64) (int, error) {
	atomic.AddInt64(&pw.written, int64(len(p)))
	percentageDownloaded := float64(pw.written) / float64(pw.size) * 100
	// int will convert 25.01 to 25.... and float64 is required Progress has float64 as required type..hence float64(int)
	pD := float64(int(percentageDownloaded))
	if math.Mod(pD, 25) == 0 {
		pw.lock.Lock()
		_, exist := pw.allowed[uint32(percentageDownloaded)]
		if exist {
			pw.progressResponse <- core.StorageManagerProgressEvent{
				Progress: pD,
			}
			delete(pw.allowed, uint32(percentageDownloaded))
		}
		pw.lock.Unlock()
	}
	return pw.writer.WriteAt(p, off)
}

// GetBlob // Generic download // COMMENT: Not used now. we are using presignedurl
func GetBlob(downloadPath string, blobCtx core.BlobContext, ctx context.Context,
	opResponse chan core.StorageManagerOperationEvent, progressResponse chan core.StorageManagerProgressEvent) error {
	s3Man := getS3ManagerMain()
	clMan := getClient()

	bucket := blobCtx.HierarchyIdentifier.Bucket

	f := log.Fields{
		"S3Provider":        s3Man.S3Option.Config.S3Provider,
		"URL":               s3Man.S3Option.Config.URL,
		"downloadPath":      downloadPath,
		"remotePathKey":     blobCtx.RemotePathKey,
		"storageIdentifier": bucket,
		"Context":           ctx,
	}

	log.WithFields(f).Info("GetBlob started")

	defer close(progressResponse)
	defer close(opResponse)

	var err error

	//Defer is scheduled at the start to be called at the end the function returns
	defer func(startTime time.Time) {
		// Error counter
		if err != nil { // parameter in label values should match the array size in metric definition
			log.WithFields(log.Fields{"error": err}).Error("GetBlob failed")
			core.StorageMetaMonitor.GetAPICounter().WithLabelValues(commonutils.GetFunctionName(GetBlob), core.MetaMonitorFAIL).Inc()
			opResponse <- core.StorageManagerOperationEvent{
				Operation: core.OperationFailure,
				Err:       err,
			}
			progressResponse <- core.StorageManagerProgressEvent{Progress: 100.0}
		} else {
			log.WithFields(log.Fields{"error": err}).Info("GetBlob successful")
			core.StorageMetaMonitor.GetAPICounter().WithLabelValues(commonutils.GetFunctionName(GetBlob), core.MetaMonitorOK).Inc()
		}

	}(time.Now())

	// to calculate the progress// COMMENT: for now not storing the process as events
	fileInfo, err := getFileInfoFromStorage(s3Man, s3Man.Ctx, blobCtx.RemotePathKey)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("failed to get file info from remote storage. check if object persist in the storage")
		return err
	}
	size := fileInfo.Size
	log.WithFields(log.Fields{"Size": size}).Info("Size of file to be downloaded")

	localPath := filepath.Dir(downloadPath)

	var retryFn = func() error {
		var err error
		err = commonutils.CreateDirectoriesIfNotExist([]string{localPath})
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("failed to create directory")
			return err
		}
		err = ioutil.WriteFile(downloadPath, []byte(""), 0644)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("failed to create file")
			return err
		}
		return err
	}

	err = commonutils.Retry(s3Man.S3ManagerConfig.RetryCount, s3Man.S3ManagerConfig.RetryDelay, retryFn,
		IsRetryableError, commonutils.ConstantBackoff)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("failed to create file")
		return err
	}

	df, err := os.OpenFile(downloadPath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("failed to open file")
		return err
	}
	defer df.Close()

	downloader := manager.NewDownloader(clMan, func(d *manager.Downloader) {
		d.Concurrency = s3Man.S3ManagerConfig.DownloadConcurrency
	})

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(blobCtx.RemotePathKey),
	}

	var allowed map[uint32]uint32

	allowed = map[uint32]uint32{25: 25, 50: 50, 75: 75, 100: 100}

	writer := &progressWriter{writer: df, size: fileInfo.Size, written: 0, progressResponse: progressResponse, allowed: allowed}

	ctx, cancel := context.WithTimeout(ctx, s3Man.S3ManagerConfig.MaxDownloadContextTimeout)
	defer cancel()

	// GetBlob the file using the AWS SDK for Go
	var retryFnAWS = func() error {
		var err error
		_, err = downloader.Download(ctx, writer, input)
		return err
	}

	err = commonutils.Retry(s3Man.S3ManagerConfig.RetryCount, s3Man.S3ManagerConfig.RetryDelay, retryFnAWS,
		IsRetryableError, commonutils.LinearBackoff)

	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("failed to download file")
		return err
	}

	metaData := fileInfo.MetaData

	i, err := strconv.ParseInt(metaData[FileModTime], 10, 64)
	if err != nil {
		panic(err)
	}
	tm := time.Unix(i, 0)

	err = os.Chtimes(downloadPath, tm, tm)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("could not modify time")
	}

	log.Info(f, "Successfully downloaded")
	opResponse <- core.StorageManagerOperationEvent{
		Operation: core.OperationSuccess,
		Err:       nil,
	}

	return nil
}

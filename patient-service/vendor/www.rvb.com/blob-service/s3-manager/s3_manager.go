package s3_manager

import (
	"context"
	"github.com/dlintw/goconf"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
	"www.rvb.com/blob-service/core"
	s3_conf "www.rvb.com/blob-service/s3-conf"
	"www.rvb.com/commonutils"
	protos "www.rvb.com/protos"
)

// Main config
type S3Manager struct {
	mux             sync.RWMutex
	Ctx             context.Context
	S3Option        *s3_conf.S3Option
	S3ManagerConfig *s3_conf.S3ManagerConfig
}

// CreateBucket : COMMENT:NOT USED : HARDCODED BUCKET FOR NOW
func (s3m *S3Manager) CreateBucket(bucketName string) error {
	_, err := CreateBucket(&bucketName)
	if err != nil {
		return err
	}
	return nil
}

func (s3m *S3Manager) PutBlob(sourcePath string, blobCtx core.BlobContext, ctx context.Context,
	opResponse chan core.StorageManagerOperationEvent, progressResponse chan core.StorageManagerProgressEvent) {
	err := PutBlob(sourcePath, blobCtx, ctx, opResponse, progressResponse)

	if err != nil {
		opResponse <- core.StorageManagerOperationEvent{
			Operation: core.OperationFailure,
			Err:       err,
		}
		progressResponse <- core.StorageManagerProgressEvent{
			Progress: 100.0}
		return
	}
}

func (s3m *S3Manager) GetBlob(downloadPath string, blobCtx core.BlobContext, ctx context.Context, opResponse chan core.StorageManagerOperationEvent, progressResponse chan core.StorageManagerProgressEvent) {
	GetBlob(downloadPath, blobCtx, ctx, opResponse, progressResponse)
}

func (s3m *S3Manager) GetBucket() string {
	if s3ManagerMain != nil {
		return s3ManagerMain.S3Option.Config.Bucket
	}
	return ""
}

// GetPresignedURL generates a presigned URL for a blob in S3.
func (s3m *S3Manager) GetPresignedURL(blobCtx core.BlobContext, ctx context.Context, expiry time.Duration, operation protos.PresignedOperation) (string, error) {
	url, err := getPresignedURL(blobCtx, ctx, expiry, operation)
	if err != nil {
		return "", err
	}
	return url, nil
}

// NewProductionStorage instantiate s3 storage,called from the client
func NewProductionStorage(c *goconf.ConfigFile) (core.StorageManager, error) {
	log.Info("NewProductionStorage , boot strapping s3 details from conf *********************************")

	err := s3ManagerInit(c)
	if err != nil {
		log.Error(log.Fields{"error": err}, "Failed to initialize s3 manager")
		return nil, err
	}

	return s3ManagerMain, nil
}

func confInit(c *goconf.ConfigFile) error {
	log.Info("confInit ****************************")
	err := s3_conf.ConfigureSections(c)
	if err != nil {
		log.Error(log.Fields{"error": err}, "failed to get conf file")
		return err
	}
	return nil
}

// Set gloabal S3 manager i,e S3ManagerMain
func s3ManagerInit(c *goconf.ConfigFile) error {

	log.Info("s3ManagerInit ****************************")

	var err error

	s3Option := &s3_conf.S3Option{Ctx: context.Background()}

	err = confInit(c)
	if err != nil {
		log.Error(log.Fields{"error": err}, "failed to get conf file")
		return err
	}

	if s3ManagerMain == nil {
		s3ManagerMain = &S3Manager{Ctx: context.Background()}
	}
	s3ManagerMain.S3ManagerConfig = s3_conf.S3ManagerCfg
	s3Option.Config = *s3_conf.S3ConfigCfg
	s3ManagerMain.S3Option = s3Option

	// Create S3 clientMain
	err = NewS3Client()
	if err != nil {
		log.Error(log.Fields{"error": err}, "failed to get clientMain object")
		return err
	}

	log.Info(log.Fields{"s3Manager": getS3ManagerMain()}, "S3Manager createddd")

	if err != nil {
		log.Error(log.Fields{"error": err}, "s3Init run env: local failed")
		return err
	}

	return nil
}

func getConfigs(c *goconf.ConfigFile, section string) (*s3_conf.S3ManagerConfig, error) {

	var err error

	s3ManagerCfg := s3_conf.S3ManagerConfig{}

	UploadConcurrency := commonutils.GetIntFromConfig(c, section, "upload_concurrency", s3_conf.DefaultUploadConcurrency)
	s3ManagerCfg.UploadConcurrency = UploadConcurrency
	log.Info(log.Fields{"UploadConcurrency": s3ManagerCfg.UploadConcurrency}, "s3Manager conf UploadConcurrency")

	DownloadConcurrency := commonutils.GetIntFromConfig(c, section, "download_concurrency", s3_conf.DefaultDownloadConcurrency)
	s3ManagerCfg.DownloadConcurrency = DownloadConcurrency
	log.Info(log.Fields{"DownloadConcurrency": s3ManagerCfg.DownloadConcurrency}, "s3Manager conf DownloadConcurrency")

	contextTimeout := commonutils.GetTimeFromConfig(c, section, "context_timeout", s3_conf.DefaultContextTimeout)
	s3ManagerCfg.ContextTimeout = contextTimeout
	log.Info(log.Fields{"ContextTimoeut": s3ManagerCfg.ContextTimeout}, "s3Manager conf ContextTimeout")

	maxUploadContextTimeout := commonutils.GetTimeFromConfig(c, section, "max_upload_context_timeout", s3_conf.DefaultMaxUploadContextTimeout)
	s3ManagerCfg.MaxUploadContextTimeout = maxUploadContextTimeout
	log.Info(log.Fields{"MaxUploadContextTimeout": s3ManagerCfg.MaxUploadContextTimeout}, "s3Manager conf MaxUploadContextTimeout")

	maxDownloadContextTimeout := commonutils.GetTimeFromConfig(c, section, "max_download_context_timeout", s3_conf.DefaultMaxDownloadContextTimeout)
	s3ManagerCfg.MaxDownloadContextTimeout = maxDownloadContextTimeout
	log.Info(log.Fields{"MaxDownloadContextTimeout": s3ManagerCfg.MaxDownloadContextTimeout}, "s3Manager conf MaxDownloadContextTimeout")

	retryDelay := commonutils.GetTimeFromConfig(c, section, "retry_delay", s3_conf.DefaultRetryDelay)
	s3ManagerCfg.RetryDelay = retryDelay
	log.Info(log.Fields{"RetryDelay": s3ManagerCfg.MaxDownloadContextTimeout}, "s3Manager conf RetryDelay")

	retryCount := commonutils.GetIntFromConfig(c, section, "retry_count", s3_conf.DefaultRetryCount)
	s3ManagerCfg.RetryCount = retryCount
	log.Info(log.Fields{"RetryCount": s3ManagerCfg.MaxDownloadContextTimeout}, "s3Manager conf RetryCount")

	expirePeriod := commonutils.GetIntFromConfig(c, section, "expire_period", s3_conf.DefaultExpireDays)
	s3ManagerCfg.ExpirePeriod = expirePeriod
	log.Info(log.Fields{"ExpirePeriod": s3ManagerCfg.ExpirePeriod}, "s3Manager conf ExpirePeriod")

	s3ManagerCfg.S3StorageClass, err = c.GetString(section, "storage_class")
	if err != nil {
		s3ManagerCfg.S3StorageClass = s3_conf.DefaultStorageClass
		log.Error(log.Fields{"error": err, "S3StorageClass": s3ManagerCfg.S3StorageClass}, "assigning default field")
	} else {
		log.Info(log.Fields{"S3StorageClass": s3ManagerCfg.S3StorageClass}, "s3Manager conf storage_class")
	}

	return &s3ManagerCfg, nil
}

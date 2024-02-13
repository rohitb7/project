package s3_conf

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/dlintw/goconf"
	log "github.com/sirupsen/logrus"
	"reflect"
	"runtime"
	"time"
	"www.rvb.com/commonutils"
)

const (
	AWSCloud                  = "AWSCloud"
	OnPrem                    = "OnPrem"
	DefaultPartSizeInBytes    = 5 * 1000 * 1000 // 5 MB
	DefaultAccelerateEndpoint = false
	DefaultListChunk          = 20
	DefaultContextTimeout     = 40 * time.Second
	//TODO : to be dynamic based on file size
	DefaultMaxUploadContextTimeout = 40 * time.Second
	//TODO : to be dynamic based on file size
	DefaultMaxDownloadContextTimeout        = 40 * time.Second
	DefaultRetryCount                       = 3
	DefaultRetryDelay                       = 3 * time.Second
	DefaultExpireDays                       = 365
	DefaultStorageClass                     = "STANDARD"
	DefaultMultipartUploadMaxRetry          = 3
	DefaultUploadConcurrency                = 5
	DefaultDownloadConcurrency              = 5
	DefaultMaxUploadContextTimeoutMultiPart = 60 * time.Second
)

type NodeConfiguration interface {
	configure(sectionName string, c *goconf.ConfigFile) error
}

func s3configConfigurationNew() NodeConfiguration {
	return &S3Config{}
}

func testConfigurationNew() NodeConfiguration {
	return &TestConfig{}
}

func s3ManagerConfigurationNew() NodeConfiguration {
	return &S3ManagerConfig{}
}

var (
	S3ConfigCfg  *S3Config
	S3ManagerCfg *S3ManagerConfig
	TestCfg      *TestConfig
)

var (
	configurationConfModule = map[string]func() NodeConfiguration{
		"s3_config":         s3configConfigurationNew,
		"s3_manager_config": s3ManagerConfigurationNew,
	}
)

var (
	configurationConfModuleTest = map[string]func() NodeConfiguration{
		"test": testConfigurationNew,
	}
)

var MANDATORY_CONF_SECTIONS = []string{
	"s3_config",
	"s3_manager_config",
}

var MANDATORY_CONF_SECTIONS_TEST = []string{
	"test",
}

type S3Config struct {
	S3Provider   string `json:"S3Provider"`
	URL          string `json:"URL"`
	AccessKey    string `json:"AccessKey"`
	AccessSecret string `json:"AccessSecret"`
	Region       string
	Bucket       string `json:"Bucket"`
	Mode         string `json:"Mode"`
}

type S3Option struct {
	Ctx    context.Context
	Config S3Config
	AwsCfg *aws.Config
}

type S3ManagerConfig struct {
	UploadConcurrency         int
	DownloadConcurrency       int
	S3StorageClass            string
	ContextTimeout            time.Duration
	MaxUploadContextTimeout   time.Duration
	MaxDownloadContextTimeout time.Duration
	RetryDelay                time.Duration
	RetryCount                int
	ExpirePeriod              int
}

type TestConfig struct {
	TestLocation              string
	TestUploadsLocation       string
	TestDownloadsLocation     string
	TestUploadsLocationPrefix string
}

// Configure each section from conf file, return error if any sections are not properly configured
func ConfigureSections(c *goconf.ConfigFile) error {
	var configErr error = nil
	var err error
	if c == nil {
		return fmt.Errorf("ConfigureSections: ConfigFile is nil")
	}

	for _, section := range MANDATORY_CONF_SECTIONS {
		log.Info(log.Fields{"section": section}, "Configuring section")
		if funcConfig, ok := configurationConfModule[section]; ok {
			err := funcConfig().configure(section, c)
			if err != nil {
				log.Error(log.Fields{"error": err, "section": section}, "Configuring section failed")
				configErr = err
			}
		} else {
			log.Info(log.Fields{"error": err, "section": section}, "section configuration found in config module, but not configurable")
		}
	}

	if configErr != nil {
		log.Error(log.Fields{"error": err, "configErr": configErr}, "Configuring processing failed")
	} else {
		log.Info("Mandatory config sections processed successfully")
	}

	return configErr
}

func ConfigureTestSections(c *goconf.ConfigFile) error {
	var configErr error = nil
	var err error
	if c == nil {
		return fmt.Errorf("ConfigureTestSections: ConfigFile is nil")
	}

	for _, section := range MANDATORY_CONF_SECTIONS_TEST {
		log.Info(log.Fields{"section": section}, "Configuring section")
		if funcConfig, ok := configurationConfModuleTest[section]; ok {
			err := funcConfig().configure(section, c)
			if err != nil {
				log.Error(log.Fields{"error": err, "section": section}, "Configuring section failed")
				configErr = err
			}
		} else {
			log.Info(log.Fields{"error": err, "section": section}, "section configuration found in config module, but not configurable")
		}
	}

	if configErr != nil {
		log.Error(log.Fields{"error": err, "configErr": configErr}, "Configuring processing failed")
	} else {
		log.Info("Mandatory config sections processed successfully")
	}

	return configErr
}

// GetFunctionName Get a function Name given a function
func GetFunctionName(function interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name()
}

// s3 configuration
func (s3Config *S3Config) configure(section string, c *goconf.ConfigFile) error {
	var err error

	s3Config.Mode, err = c.GetString(section, "mode")
	if err != nil {
		log.Error(log.Fields{"error": err, "mode": s3Config.Mode}, "missing conf field")
		return err
	} else {
		log.Info(log.Fields{"Mode": s3Config.Mode}, "conf mode")
	}

	s3Config.Bucket, err = c.GetString(section, "bucket")
	if err != nil {
		log.Error(log.Fields{"error": err, "bucket": s3Config.Bucket}, "missing conf field")
		return err
	} else {
		log.Info(log.Fields{"Bucket": s3Config.Bucket}, "bucket conf Bucket")
	}

	s3Config.Region, err = c.GetString(section, "region")
	if err != nil {
		log.Error(log.Fields{"error": err, "Region": s3Config.Region}, "missing conf field")
		return err
	} else {
		log.Info(log.Fields{"Region": s3Config.Region}, "conf region")
	}

	s3Config.S3Provider, err = c.GetString(section, "s3_provider")
	if err != nil {
		log.Error(log.Fields{"error": err, "S3Provider": s3Config.S3Provider}, "missing conf field")
		return err
	} else {
		log.Info(log.Fields{"S3Provider": s3Config.S3Provider}, "s3 conf S3Provider")
	}

	s3Config.URL, err = c.GetString(section, "url")
	if err != nil {
		log.Error(log.Fields{"error": err, "URL": s3Config.URL}, "missing conf field")
		return err
	} else {
		log.Info(log.Fields{"URL": s3Config.URL}, "s3 conf URL")
	}

	s3Config.AccessKey, err = c.GetString(section, "access_key")
	if err != nil {
		log.Error(log.Fields{"error": err, "AccessKey": s3Config.AccessKey}, "missing conf field")
		return err
	} else {
		log.Info(log.Fields{"AccessKey": s3Config.AccessKey}, "s3 conf AccessKey")
	}

	s3Config.AccessSecret, err = c.GetString(section, "secret_access")
	if err != nil {
		log.Error(log.Fields{"error": err, "bucket": s3Config.AccessSecret}, "missing conf field")
		return err
	} else {
		log.Info(log.Fields{"AccessSecret": s3Config.AccessSecret}, "s3 conf AccessSecret")
	}

	S3ConfigCfg = s3Config
	return nil
}

// s3_manager  configuration
func (s3ManagerCfg *S3ManagerConfig) configure(section string, c *goconf.ConfigFile) error {
	var err error

	UploadConcurrency := commonutils.GetIntFromConfig(c, section, "upload_concurrency", DefaultUploadConcurrency)
	s3ManagerCfg.UploadConcurrency = UploadConcurrency
	log.Info(log.Fields{"UploadConcurrency": s3ManagerCfg.UploadConcurrency}, "s3Manager conf UploadConcurrency")

	DownloadConcurrency := commonutils.GetIntFromConfig(c, section, "download_concurrency", DefaultDownloadConcurrency)
	s3ManagerCfg.DownloadConcurrency = DownloadConcurrency
	log.Info(log.Fields{"DownloadConcurrency": s3ManagerCfg.DownloadConcurrency}, "s3Manager conf DownloadConcurrency")

	contextTimeout := commonutils.GetTimeFromConfig(c, section, "context_timeout", DefaultContextTimeout)
	s3ManagerCfg.ContextTimeout = contextTimeout
	log.Info(log.Fields{"ContextTimoeut": s3ManagerCfg.ContextTimeout}, "s3Manager conf ContextTimoeut")

	maxUploadContextTimeout := commonutils.GetTimeFromConfig(c, section, "max_upload_context_timeout", DefaultMaxUploadContextTimeout)
	s3ManagerCfg.MaxUploadContextTimeout = maxUploadContextTimeout
	log.Info(log.Fields{"MaxUploadContextTimeout": s3ManagerCfg.MaxUploadContextTimeout}, "s3Manager conf MaxUploadContextTimeout")

	maxDownloadContextTimout := commonutils.GetTimeFromConfig(c, section, "max_download_context_timeout", DefaultMaxDownloadContextTimeout)
	s3ManagerCfg.MaxDownloadContextTimeout = maxDownloadContextTimout
	log.Info(log.Fields{"MaxDownloadContextTimeout": s3ManagerCfg.MaxDownloadContextTimeout}, "s3Manager conf MaxDownloadContextTimeout")

	retryDelay := commonutils.GetTimeFromConfig(c, section, "retry_delay", DefaultRetryDelay)
	s3ManagerCfg.RetryDelay = retryDelay
	log.Info(log.Fields{"RetryDelay": s3ManagerCfg.MaxDownloadContextTimeout}, "s3Manager conf RetryDelay")

	retryCount := commonutils.GetIntFromConfig(c, section, "retry_count", DefaultRetryCount)
	s3ManagerCfg.RetryCount = retryCount
	log.Info(log.Fields{"RetryCount": s3ManagerCfg.MaxDownloadContextTimeout}, "s3Manager conf RetryCount")

	expirePeriod := commonutils.GetIntFromConfig(c, section, "expire_period", DefaultExpireDays)
	s3ManagerCfg.ExpirePeriod = expirePeriod
	log.Info(log.Fields{"ExpirePeriod": s3ManagerCfg.ExpirePeriod}, "s3Manager conf ExpirePeriod")

	s3ManagerCfg.S3StorageClass, err = c.GetString(section, "storage_class")
	if err != nil {
		s3ManagerCfg.S3StorageClass = DefaultStorageClass
		log.Error(log.Fields{"error": err, "S3StorageClass": s3ManagerCfg.S3StorageClass}, "assigning default field")
	} else {
		log.Info(log.Fields{"S3StorageClass": s3ManagerCfg.S3StorageClass}, "s3Manager conf storage_class")
	}

	S3ManagerCfg = s3ManagerCfg
	return nil
}

// s3_test  configuration
func (testConfig *TestConfig) configure(section string, c *goconf.ConfigFile) error {
	var err error

	testConfig.TestLocation, err = c.GetString(section, "test_dir")
	if err != nil {
		log.Error(log.Fields{"error": err, "TestLocation": testConfig.TestLocation}, "missing conf field")
		return err
	} else {
		log.Info(log.Fields{"TestLocation": testConfig.TestLocation}, "TestConfig conf TestLocation")
	}

	testConfig.TestUploadsLocation, err = c.GetString(section, "test_uploads_location")
	if err != nil {
		log.Error(log.Fields{"error": err, "TestUploadsLocation": testConfig.TestUploadsLocation}, "missing conf field")
		return err
	} else {
		log.Info(log.Fields{"TestUploadsLocation": testConfig.TestUploadsLocation}, "TestConfig conf TestUploadsLocation")
	}

	testConfig.TestDownloadsLocation, err = c.GetString(section, "test_downloads_location")
	if err != nil {
		log.Error(log.Fields{"error": err, "TestDownloadsLocation": testConfig.TestDownloadsLocation}, "missing conf field")
		return err
	} else {
		log.Info(log.Fields{"TestDownloadsLocation": testConfig.TestDownloadsLocation}, "TestConfig conf TestDownloadsLocation")
	}

	TestCfg = testConfig
	return nil
}

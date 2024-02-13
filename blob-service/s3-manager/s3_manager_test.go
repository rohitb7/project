package s3_manager

import (
	"context"
	"github.com/dlintw/goconf"
	log "github.com/sirupsen/logrus"
	"strings"
	"testing"
	"www.rvb.com/blob-service/s3-conf"
)

var (
	configFile = "./../s3-conf/s3_conf_test.conf"
)

// confInitTest will test the configurations only
func confInitTest() error {
	log.Info("confInitTest **************************************************")
	c, err := goconf.ReadConfigFile(configFile)
	if err != nil {
		// log.Error("confInitTest: ReadConfigFile failed: %v", err)
		return err
	}
	// Mandatory Sections
	err = s3_conf.ConfigureSections(c)
	if err != nil {
		// log.ErrorF(// log.Fields{"error": err}, "failed to get conf file")
		return err
	}
	// Test Sections needed to init tests needed configs
	err = s3_conf.ConfigureTestSections(c)
	if err != nil {
		// log.ErrorF(// log.Fields{"error": err}, "failed to get conf file")
		return err
	}
	return nil
}

// s3ManagerInitTest will test the configurations only and create a s3Manager which serves as a parent struct for other tests
func s3ManagerInitTest() error {

	log.Info("s3ManagerInitTest")
	s3OptionTest := &s3_conf.S3Option{Ctx: context.Background()}

	err := confInitTest()
	if err != nil {
		// log.ErrorF(// log.Fields{"error": err}, "failed to get conf file")
		return err
	}

	// Define a fresh s3 manager
	s3ManagerMain = &S3Manager{Ctx: context.Background()}
	s3ManagerMain.S3ManagerConfig = s3_conf.S3ManagerCfg

	s3OptionTest.Config = *s3_conf.S3ConfigCfg
	s3ManagerMain.S3Option = s3OptionTest

	log.Info("S3Manager created")

	// log.InfoF(// log.Fields{"s3ManagerTest": getS3ManagerMain()}, "Test S3Manager")

	if err != nil {
		// log.ErrorF(// log.Fields{"error": err}, "s3Init run env: local failed")
		return err
	}
	return nil
}

// testS3ManagerInit will test the complete configurations and create a s3Manager and creation of s3 client as well
func testS3ManagerInit(t *testing.T) {
	var err error
	err = s3ManagerInitTest()
	if err != nil {
		t.Fatalf("Failed to int s3 manager")
	}

	// Create S3 clientMain
	err = NewS3Client()
	if err != nil {
		t.Fatalf("Failed to create S3 cliant")
	}

	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			t.Fatalf("Skip error: connection refused")
		} else {
			t.Fatalf("testS3ManagerInit() %v", err)
		}
	}
}

// s3TestLocationsInit will initialize temporary  locations for test upload and downloads
func s3TestLocationsInit() (*s3_conf.TestConfig, error) {
	log.Info("s3TestLocationsInit")
	err := confInitTest()
	if err != nil {
		return nil, err
	}
	testConfig := &s3_conf.TestConfig{
		TestLocation:              s3_conf.TestCfg.TestLocation,
		TestUploadsLocationPrefix: s3_conf.TestCfg.TestUploadsLocationPrefix,
		TestUploadsLocation:       s3_conf.TestCfg.TestUploadsLocation,
		TestDownloadsLocation:     s3_conf.TestCfg.TestDownloadsLocation,
	}

	log.Info("s3Test created", testConfig)
	return testConfig, nil
}

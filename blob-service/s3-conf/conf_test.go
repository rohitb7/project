package s3_conf

import (
	"github.com/dlintw/goconf"
	log "github.com/sirupsen/logrus"
	"testing"
	"www.rvb.com/commonutils"
)

var (
	configFile = "s3_conf_test.conf"
)

func testCommonConfServicesSetup(t *testing.T, section string) {
	var downloadsLocation, uploadsLocation, testFolderLocation string
	c, err := goconf.ReadConfigFile(configFile)
	if err != nil {
		log.Error(log.Fields{"error": err}, "Config Read failed ")
		t.Fatalf("Config read failed %+v", err)
	}

	testFolderLocation, err = c.GetString(section, "test_dir")
	if err != nil {
		log.Error(log.Fields{"error": err, "section": section}, "missing test folder location")
		t.Fatalf("failed to get dir, error %v", err)
	}

	downloadsLocation, err = c.GetString(section, "test_downloads_location")
	if err != nil {
		log.Error(log.Fields{"error": err, "section": section}, "missing test downloads location")
		t.Fatalf("failed to get dir, error %v", err)
	}

	uploadsLocation, err = c.GetString(section, "test_uploads_location")
	if err != nil {
		log.Error(log.Fields{"error": err, "section": section}, "missing test uploads location")
		t.Fatalf("failed to get dir, error %v", err)
	}

	commonutils.CreateDirIfNotExist(testFolderLocation)
	commonutils.CreateDirIfNotExist(uploadsLocation)
	commonutils.CreateDirIfNotExist(downloadsLocation)

	log.Info(log.Fields{
		"test_location":      testFolderLocation,
		"uploads_location":   uploadsLocation,
		"downloads_location": downloadsLocation,
	},
		"configured dir location")

	err = commonutils.RemovePaths([]string{testFolderLocation, uploadsLocation, downloadsLocation})
	if err != nil {
		log.Error(log.Fields{"error": err, "section": section}, "failed to remove paths")
		t.Fatalf("failed to remove paths, error %v", err)
	}

	log.Info(log.Fields{
		"test_location":      testFolderLocation,
		"uploads_location":   uploadsLocation,
		"downloads_location": downloadsLocation,
	},
		"removed dir location")
}

// Common Config Setup and TearDown
func TestCommonAllConfigSetup(t *testing.T) {
	log.Info(log.Fields{"Test": commonutils.GetTestName(t)}, "****** Started ******")
	testCommonConfServicesSetup(t, "test")
	log.Info("Common Services Setup Success")
	log.Info(log.Fields{"Test": commonutils.GetTestName(t)}, "****** Ended ******")
}

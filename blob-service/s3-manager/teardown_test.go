package s3_manager

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
	"time"
	"www.rvb.com/commonutils"
)

// file for upload
const fileName = "temp-file"

func TestTearDown(t *testing.T) {
	testCases := []testFn{
		testS3ManagerInit,
		//testCreateBucket, TODO facing issues with minio. not consistent
		teardown,
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Running %s ", commonutils.GetFunctionName(tc)), tc)
	}
}

func teardown(t *testing.T) {
	//TODO: args can be passed, currently unnecessary re-configuration is happening
	testInit, err := s3TestLocationsInit()
	if err != nil {
		log.WithFields(log.Fields{"err": err.Error()}).Info("Could not init test configs")
		t.Fatalf("Could not init test configs")
	}
	os.RemoveAll(testInit.TestLocation)
	time.Sleep(1 * time.Second)
}

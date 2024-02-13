package s3_manager

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
	"testing"
	"www.rvb.com/commonutils"
)

func TestCreateBucket(t *testing.T) {
	testCases := []testFn{
		testS3ManagerInit,
		testCreateBucket,
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Running %s ", commonutils.GetFunctionName(tc)), tc)
	}
	teardown(t)
}

func testCreateBucket(t *testing.T) {
	log.WithFields(log.Fields{"Test": commonutils.GetTestName(t)}).Info("****** Started ******")
	var err error
	_, err = CreateBucket(nil)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			t.Fatalf("Skip error: connection refused")
		} else {
			t.Fatalf("CreateBucket() %v", err)
		}
	}
	log.WithFields(log.Fields{"Test": commonutils.GetTestName(t)}).Info("****** Ended ******")
}

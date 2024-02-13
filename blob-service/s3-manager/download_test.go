package s3_manager

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
	"www.rvb.com/blob-service/core"
	"www.rvb.com/commonutils"
)

// TestDownload 1)upload 2)download 3)teardown
func TestDownload(t *testing.T) {
	testCases := []testFn{
		testS3ManagerInit,
		//testCreateBucket,
		testUpload,
		testDownloadSingle,
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Running %s ", commonutils.GetFunctionName(tc)), tc)
	}

	teardown(t)
}

func testDownloadSingle(t *testing.T) {

	testInit, err := s3TestLocationsInit()
	if err != nil {
		log.WithFields(log.Fields{"err": err.Error()}).Error("Could not init test configs")
		t.Fatalf("Could not init test configs")
	}

	err = commonutils.CreateDirIfNotExist(testInit.TestDownloadsLocation)
	if err != nil {
		return
	}

	var Tests = []struct {
		testName      string
		withCancelCtx bool
		fileName      string
		isErrExpected bool
	}{
		{
			testName:      "WithoutCancelContext",
			withCancelCtx: false,
			fileName:      fileName,
			isErrExpected: false,
		},
		{
			testName:      "WithCancelContext",
			withCancelCtx: true,
			fileName:      fileName,
			isErrExpected: true,
		},
	}

	for _, tc := range Tests {
		t.Run(tc.testName, func(t *testing.T) {
			downloadPath := testInit.TestDownloadsLocation + "/" + tc.fileName
			testDownloadHelper(t, tc.fileName, tc.withCancelCtx, tc.isErrExpected, downloadPath)
		})
		time.Sleep(1 * time.Second)
	}
}

func testDownloadHelper(t *testing.T, fileName string, withCancelCtx bool, isErrExpected bool, downloadPath string) {

	log.WithFields(log.Fields{"Test": commonutils.GetTestName(t)}).Info("****** Started ******")
	var err error

	ctx := context.Background()
	if withCancelCtx {
		cancelFunc := func() {}
		ctx, cancelFunc = context.WithTimeout(context.Background(), 5*time.Second)
		cancelFunc()
	}

	remotePathKey := fileName
	storageIdentifier := core.HierarchyIdentifier{Bucket: getS3ManagerMain().S3Option.Config.Bucket}
	blobCtx := core.BlobContext{
		RemotePathKey:       remotePathKey,
		HierarchyIdentifier: storageIdentifier,
	}

	fileInfo, err := getFileInfoFromStorage(getS3ManagerMain(), ctx, fileName)
	if isErrExpected && err != nil {
		t.SkipNow()
	}
	if !isErrExpected && err != nil {
		t.Fatalf("not able to get fileInfo")
	}

	var progressResponse = make(chan core.StorageManagerProgressEvent, 100)
	var opResponse = make(chan core.StorageManagerOperationEvent, 1)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				log.WithFields(log.Fields{"error": ctx.Err(), "remotePathKey": remotePathKey}).Error("Downloaded.")
				return
			case progress, ok := <-progressResponse:
				log.Info(log.Fields{"progress": progress, "ok": ok})
				if !ok {
					progressResponse = nil
				}
			case progress, ok := <-opResponse:
				log.Info(log.Fields{"progress": progress, "ok": ok})
				if !ok {
					opResponse = nil
				}
			}
			if opResponse == nil || progressResponse == nil {
				break
			}
		}
	}()

	err = GetBlob(downloadPath, blobCtx, ctx, opResponse, progressResponse)

	downloadedFileInfo, err := os.Stat(downloadPath)
	if err != nil {
		t.Fatalf("file not found")
	}

	if fileInfo.MetaData[FileModTime] != fmt.Sprint(downloadedFileInfo.ModTime().Unix()) {
		t.Fatalf("mod time did not matched")
	}

	if isErrExpected && err == nil {
		t.Fatalf("Expected error, Got no error")
	}

	if isErrExpected && err != nil {
		log.WithFields(log.Fields{"err": err.Error()}).Error("Context cancelled error")
		t.SkipNow()
	}

	if !isErrExpected && err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			t.Fatalf("Skip error: connection refused")
		} else if strings.Contains(err.Error(), "NoSuchBucket") {
			t.Skip("Skip error: NoSuchBucket")
		} else if strings.Contains(err.Error(), "NoSuchKey") {
			t.Skip("Skip error: NoSuchKey")
		} else {
			t.Fatalf("GetBlob() %v", err)

		}
	}
	log.WithFields(log.Fields{"Test": commonutils.GetTestName(t)}).Error("****** Ended ******")
}

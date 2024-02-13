package s3_manager

import (
	"context"
	"fmt"
	_ "github.com/prometheus/client_model/go"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
	"www.rvb.com/blob-service/core"
	"www.rvb.com/commonutils"
)

// COMMENT: this test runs at the bootup. uploads all images files
func TestUploadAllFiles(t *testing.T) {

	testS3ManagerInit(t)

	_, err := s3TestLocationsInit()
	if err != nil {
		t.Fatalf("Could not init test configs: %v", err)
	}

	//hardcoding to cd into the images directory and upload all the images
	workspaceDir := os.Getenv("WORKSPACE_DIR")
	if workspaceDir == "" {
		fmt.Println("The WORKSPACE_DIR environment variable is not set.")
	} else {
		fmt.Printf("The WORKSPACE_DIR is set to: %s\n", workspaceDir)
	}

	uploadLocation := workspaceDir + "/temperory-list-of-images"

	files, err := ioutil.ReadDir(uploadLocation)
	if err != nil {
		t.Fatalf("Could not read directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".jpg") || strings.HasSuffix(file.Name(), ".jpeg") { // Ensures only .jpg files are uploaded
			t.Run(file.Name(), func(t *testing.T) {
				originalFilePath := filepath.Join(uploadLocation, file.Name())
				tempFilePath := filepath.Join(".", filepath.Base(originalFilePath)) // Temporary file in the current directory

				// Create a temporary copy of the file in the current directory
				if err := commonutils.CopyFile(originalFilePath, tempFilePath); err != nil {
					t.Fatalf("Failed to create a temporary copy of the file: %v", err)
				}

				// Ensure the temporary file is deleted after the upload
				defer os.Remove(tempFilePath)

				blobCtx := core.BlobContext{
					RemotePathKey: filepath.Base(tempFilePath),
					HierarchyIdentifier: core.HierarchyIdentifier{
						Bucket: "mybucket",
					},
				}

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				var progressResponse = make(chan core.StorageManagerProgressEvent, 100)
				var opResponse = make(chan core.StorageManagerOperationEvent, 1)

				wg := sync.WaitGroup{}
				wg.Add(1)
				go func() {
					defer wg.Done()
					for {
						select {
						case <-ctx.Done():
							log.WithFields(log.Fields{"error": ctx.Err(), "filepath.Base(tempFilePath)": filepath.Base(tempFilePath)}).Info("Image uploaded.")
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
							log.Info("BREAK")
							break
						}
					}
				}()

				err := PutBlob(filepath.Base(tempFilePath), blobCtx, ctx, opResponse, progressResponse)
				if err != nil {
					t.Errorf("Failed to upload %s: %v", file.Name(), err)
				}
			})
		}
	}
}

func TestUpload(t *testing.T) {
	testCases := []testFn{
		testS3ManagerInit,
		//testCreateBucket,
		testUpload,
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Running %s ", commonutils.GetFunctionName(tc)), tc)
	}
	teardown(t)
}

// Single PutBlob
func testUpload(t *testing.T) {

	testInit, err := s3TestLocationsInit()
	if err != nil {
		log.WithFields(log.Fields{"err": err.Error()}).Error("Could not init test configs")
		t.Fatalf("Could not init test configs")
	}

	uploadLocation := testInit.TestUploadsLocation

	var Tests = []struct {
		testName      string
		withCancelCtx bool
		fileSize      int64
		isErrExpected bool
	}{
		{
			testName:      "WithoutCancelContext",
			withCancelCtx: false,
			fileSize:      2,
			isErrExpected: false,
		},
		{
			testName:      "WithCancelContext",
			withCancelCtx: true,
			fileSize:      2,
			isErrExpected: true,
		},
	}

	for _, tc := range Tests {
		t.Run(tc.testName, func(t *testing.T) {
			testUploadHelper(t, tc.fileSize, tc.withCancelCtx, tc.isErrExpected, uploadLocation, false)
		})
		time.Sleep(1 * time.Second)
	}
}

func testUploadHelper(t *testing.T, fileSize int64, withCancelCtx bool, isErrExpected bool, uploadLocation string, withPrefix bool) {

	log.WithFields(log.Fields{"Test": commonutils.GetTestName(t)}).Info("****** Started ******")

	var err error

	testInit, err := s3TestLocationsInit()
	if err != nil {
		log.WithFields(log.Fields{"err": err.Error()}).Error("Could not init test configs")
		t.Fatalf("Could not init test configs")
	}

	ctx := context.Background()
	if withCancelCtx {
		cancelFunc := func() {}
		ctx, cancelFunc = context.WithTimeout(context.Background(), 2*time.Second)
		cancelFunc()
	}

	fileName := fileName

	err = commonutils.CreateTempFile(fileSize, fileName, uploadLocation)
	if err != nil {
		t.Fatalf("TestUpload() %v", err)
	}

	sourcePath := filepath.Join(uploadLocation, fileName)

	remotePathKey := ""
	if withPrefix {
		msec := time.Now().UnixNano() / 1000000
		remotePathKey = testInit.TestUploadsLocationPrefix + "/" + fileName + fmt.Sprint(msec)
	} else {
		remotePathKey = fileName
	}

	storageIdentifier := core.HierarchyIdentifier{Bucket: getS3ManagerMain().S3Option.Config.Bucket}
	blobCtx := core.BlobContext{
		RemotePathKey:       remotePathKey,
		HierarchyIdentifier: storageIdentifier,
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
				log.WithFields(log.Fields{"error": ctx.Err(), "remotePathKey": remotePathKey}).Error("Upload Successfull")
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
				log.Info("BREAK")
				break
			}
		}
	}()

	err = PutBlob(sourcePath, blobCtx, ctx, opResponse, progressResponse)

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
		} else {
			t.Fatalf("PutBlob() %v", err)
		}
	}

	log.WithFields(log.Fields{"Test": commonutils.GetTestName(t)}).Error("****** Ended ******")
}

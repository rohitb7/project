package main

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
	"www.rvb.com/blob-service/core"
	protoenumutils "www.rvb.com/patient-service/protoenumutils"
	protos "www.rvb.com/protos"
)

var (
	ErrFailedToConnectToDatabase = errors.New("failed to connect to database")
	ErrQueryingImagesTable       = errors.New("error querying images table")
	ErrIteratingOverImageRows    = errors.New("error iterating over image rows")
	ErrScanningImageRow          = errors.New("error scanning image row")
	ErrJobQueueFull              = errors.New("job is queue is full")
	ErrEmptyProtoRequest         = errors.New("empty proto request")
	ErrNoPatientIdInReq          = errors.New("no patient id in request")
	ErrFileSizeExcedded          = errors.New("fileSize Excedded")
	ErrBucketAdded               = errors.New("bucket change in storage provider request without force flag")
)

// handleRetrievePatientImage
// Synchronous call
// this method stores data in db and send a presignedurl
func handleRetrievePatientImage(protoRequest *protos.ListPatientImagesRequest, ctx context.Context) *protos.ListPatientImagesResponse {

	log.Info("handleRetrievePatientImage started")

	var err error

	protoResponse := &protos.ListPatientImagesResponse{
		Patient: &protos.Patient{
			Id:       "",
			Name:     "",
			UserName: "",
		},
		Images: nil,
		Result: &protos.Result{
			RequestResult: 0,
			Error: &protos.Error{
				Message:          "",
				ErrorRetryStatus: 0,
			},
		},
	}

	retryStatus, err := protoenumutils.GetErrorRetryStatusEnumValueFromString(protos.ErrorRetryStatus_RETRY.String())
	if err != nil {
		return protoResponse
	}

	defer func() {
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("failed to get patient list of images")
			protoResponse.Result = &protos.Result{
				RequestResult: protos.RequestResult_REJECTED,
				Error: &protos.Error{
					Message:          err.Error(),
					ErrorRetryStatus: retryStatus,
				},
			}
		} else {
			log.Info("successfully got patient list of images")
			protoResponse.Result = &protos.Result{
				RequestResult: protos.RequestResult_ACCEPTED,
			}
		}
	}()

	if protoRequest == nil {
		err = ErrEmptyProtoRequest
		return protoResponse
	}

	if len(protoRequest.GetPatient().GetId()) == 0 {
		err = ErrNoPatientIdInReq
		return protoResponse
	}

	// adds to db
	protoResponse, err = retrievePatientImageHandlerDB(protoRequest)
	if err != nil {
		return nil
	}

	// code for concurrent go routines to get list of images from the s3/minio
	wg := sync.WaitGroup{}
	errorCh := make(chan error, len(protoResponse.Images))
	presignedURLs := make(map[string]string)
	var mu sync.Mutex
	for _, image := range protoResponse.Images {
		wg.Add(1)
		go func(img *protos.Image) {
			defer wg.Done()
			remoteFile := img.GetBucketPath() // Assuming GetPath gets the remote path key
			// Generate presigned URL
			url, err := serverCfg.storageManagerInterface.GetPresignedURL(core.BlobContext{
				RemotePathKey: remoteFile,
				HierarchyIdentifier: core.HierarchyIdentifier{
					Bucket: s3Cfg.bucket,
				},
			}, ctx, time.Minute*10, 0) // Presigned URL expires in 10 minutes, //TODO replace operation 0 = DOWNLOAD.
			if err != nil {
				errorCh <- err
				return
			}
			mu.Lock()
			presignedURLs[remoteFile] = url
			mu.Unlock()
		}(image)
	}
	wg.Wait()
	// Close the error channel and handle any errors
	close(errorCh)
	for err = range errorCh {
		if err != nil {
			// Handle the error, e.g., log it or return it? depends as per use case. some images might return error scenario?
			log.WithFields(log.Fields{"error": err}).Error("Error generating presigned URL:")
			return protoResponse
		}
	}

	// Assign the presigned URLs back to the protoResponse
	for _, image := range protoResponse.Images {
		remoteFile := image.GetBucketPath()
		if url, ok := presignedURLs[remoteFile]; ok {
			image.Url = url
			//BucketPath should not be exposed to the UI, looks like protobuf empty object behaviour has changed
			image.BucketPath = ""
		}
	}
	log.Info("got all the image urls from s3")
	return protoResponse
}

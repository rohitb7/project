package main

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
	"www.rvb.com/blob-service/core"
	"www.rvb.com/commonutils"
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

// handleUploadPatientImage
// Asynchronous call
// this method stores data in db, and uploads file asynchronously.
// it also keeps tracks of upload status and its keeps track of its progress (currently logging but not stored in db)
// after receiving status it sets
func handleUploadPatientImageHttp(w http.ResponseWriter, r *http.Request) {
	// Parse the multipart form data
	err := r.ParseMultipartForm(10 << 20) // 10 MB maximum file size
	if err != nil {
		log.Error("Error parsing form data")
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Retrieve the file from the form data
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Error("Error retrieving file from form data")
		http.Error(w, "Unable to retrieve file from form data", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read the payload from the request body
	payload := r.FormValue("payload")
	if payload == "" {
		log.Error("Payload is empty")
		http.Error(w, "Payload is empty", http.StatusBadRequest)
		return
	}

	filePath := fmt.Sprintf("/var/tmp/uploads/%s", handler.Filename)
	newFile, err := os.Create(filePath)
	if err != nil {
		log.Error("Error creating file on serve")
		http.Error(w, "Unable to create file on server", http.StatusInternalServerError)
		return
	}
	defer newFile.Close()

	// Copy the uploaded file data to the newly created file on the server
	_, err = io.Copy(newFile, file)
	if err != nil {
		log.Error("Error writing file data to disk")
		http.Error(w, "Unable to write file data to disk", http.StatusInternalServerError)
		return
	}

	// todo: fix : hardcoded
	protoRequest := &protos.UploadPatientImageRequest{
		PatientImage: &protos.PatientImage{
			PatientId: "1",
			Image: &protos.ImageUI{
				Name:        "",
				Description: "",
				Tags:        nil,
			},
		},
		Tags: &protos.Tags{
			Tag: nil,
		},
		FilePath: filePath,
	}

	// Respond with success message
	log.Info("File uploaded on sever, will be pushed to s3 asynchronously")
	w.WriteHeader(http.StatusOK)

	if workPoolCfg.masterPool.IsQueueFull() {
		err = ErrJobQueueFull
		w.WriteHeader(http.StatusForbidden)
	}

	//COMMENT: store each async job in db and its s tatus can be checked later
	jobId := commonutils.RandomStringUsingTimestamp(10)

	ctx, cancel := context.WithCancel(context.Background())
	t := &UploadImageImpl{
		asyncRunnerBase: asyncRunnerBase{
			JobId:         jobId,
			JobType:       protos.JobType_UPLOAD_IMAGE,
			Context:       ctx,
			CancelContext: cancel,
		},
		ProtoRequest: protoRequest,
	}
	JobManagerInstance.RegisterJob(t.GetJobId(), t, t.GetCancelFunc(), workPoolCfg.masterPool)

}

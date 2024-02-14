package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"sync"
	"www.rvb.com/blob-service/core"
	"www.rvb.com/commonutils"
	protos "www.rvb.com/protos"
)

// Async Runner interface impl
type UploadImageImpl struct {
	asyncRunnerBase
	ProtoRequest *protos.UploadPatientImageRequest
}

func (r *UploadImageImpl) GetRunnerInput() proto.Message {
	return r.ProtoRequest
}

// Async Upload
func (r *UploadImageImpl) Run() error {

	// this is a async job. which as an acknowledgement create a job id and returns to the client as soon as a job starts. the client can poll or ping for its status later.
	// since file upload can block the UI
	var jobError error

	defer func() {
		if jobError != nil {
			//remove the job from the queue and set status db error
			handleCompleteJob(r, jobError)
			log.WithFields(log.Fields{"error": jobError.Error()}).Error("job failed")
		} else {
			//remove the job from the queue and set status db as no
			handleCompleteJob(r, nil)
			log.Info("successfully uploaded file")
		}
	}()

	protoRequest, ok := r.GetRunnerInput().(*protos.UploadPatientImageRequest)
	if !ok {
		return fmt.Errorf("failed to cast %+v", r.GetRunnerInput())
	}

	// TODO: Filesize limit check. if large file then the storage service needs to make multipart upload
	// TODO: MD5 checksum

	var progressEvents = make(chan core.StorageManagerProgressEvent, 100)
	var jobEvents = make(chan core.StorageManagerOperationEvent, 1)

	// File upload starts. currently does not handle case for bulk upload
	// Spin up the manager thread to track upload status
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-r.GetCancelContext().Done():
				log.WithFields(log.Fields{"error": jobError}).Error("Job aborted")
				return
			case e, ok := <-progressEvents:
				if !ok {
					return
				}
				if e.Progress == 100 {
					return
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-r.GetCancelContext().Done():
				log.WithFields(log.Fields{"error": jobError}).Error("Job aborted")
				return
			case e, ok := <-jobEvents:
				if !ok {
					return
				}
				if e.Operation != core.OperationSuccess {
					log.WithFields(log.Fields{"error": commonutils.GetErrorMessage(e.Err)}).Error("operation failed")
					return
				} else {
					log.Info("Job success")
					return
				}
			}
		}
	}()

	go func() {
		// spin another thread for actual file transfer // serverCfg.storageManagerInterface can be a seen as a blob service which manges CRUD operations with S3
		log.Info("transferring file to remote storage")
		serverCfg.storageManagerInterface.PutBlob(protoRequest.FilePath, core.BlobContext{
			RemotePathKey: protoRequest.FilePath,
			HierarchyIdentifier: core.HierarchyIdentifier{
				Bucket: s3Cfg.bucket,
			},
		}, r.GetCancelContext(), jobEvents, progressEvents)
	}()
	wg.Wait()

	if jobError == nil {
		go commonutils.DeleteFile(protoRequest.FilePath)
		response, err := uploadPatientImageHandlerDB(protoRequest, protoRequest.FilePath)
		if err != nil {
			return nil
		}
		log.Info("successfully uploaded file to S3", response.Result.RequestResult.String())
	} else {
		log.Error("failed transferring file to remote storage")
		// COMMENTS: not implemented
		// if response is not successful then let the client know // depends upon business use case
		// we can store the status and request of the job and retry, also the client can ping or poll or websockets ideally to get the upload status
	}
	return jobError
}

package main

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"www.rvb.com/commonutils"
	protos "www.rvb.com/protos"
)

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

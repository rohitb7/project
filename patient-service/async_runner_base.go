package main

import (
	"context"
	"github.com/golang/protobuf/proto" //nolint
	log "github.com/sirupsen/logrus"
	protos "www.rvb.com/protos"
)

// Async Runner
type AsyncRunner interface {
	GetJobId() string
	Run() error
	GetRunnerInput() proto.Message
	GetJobType() protos.JobType
	GetCancelFunc() context.CancelFunc
	GetCancelContext() context.Context
}

// Common fields for Async Runner
type asyncRunnerBase struct {
	JobId         string
	JobType       protos.JobType
	Context       context.Context
	CancelContext context.CancelFunc
}

func (r *asyncRunnerBase) GetJobType() protos.JobType {
	return r.JobType
}

func (r *asyncRunnerBase) GetJobId() string {
	return r.JobId
}

func (r *asyncRunnerBase) GetCancelFunc() context.CancelFunc {
	return r.CancelContext
}

func (r *asyncRunnerBase) GetCancelContext() context.Context {
	return r.Context
}

// handleCompleteJob remove the jobmanager
func handleCompleteJob(r AsyncRunner, jobError error) {
	var err error
	if jobError != nil {
		log.WithFields(log.Fields{"error": jobError, "job_id": r.GetJobId(), "job_type": r.GetJobType()}).Error("job failed")
	}
	if err == nil {
		log.WithFields(log.Fields{"job_id": r.GetJobId(), "job_type": r.GetJobType()}).Info("Completed DssManagerJob")
	}
	JobManagerInstance.RemoveJob(r.GetJobId())
}

package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"sync"
	"www.rvb.com/patient-service/workerpool"
)

var JobManagerInstance *JobManager

func initJobManger() {
	log.Info(log.Fields{"name": "global job manager"}, "JobManager init...")
	JobManagerInstance = &JobManager{
		jobLock:    sync.RWMutex{},
		jobContext: make(map[string]context.CancelFunc),
		jobs:       make(chan Job, 2048+1024),
		cancelJobs: make(chan bool),
		quit:       make(chan bool),
	}
}

type JobType int

const (
	RegisterJob JobType = iota
	RemoveJob
	CancelJob
)

type Job struct {
	jobId        string
	task         workerpool.Task
	ctx          context.CancelFunc
	workPoolType *workerpool.TaskManager
	jobType      JobType
}

type JobManager struct {
	jobLock    sync.RWMutex
	jobContext map[string]context.CancelFunc
	jobs       chan Job
	cancelJobs chan bool
	quit       chan bool
}

// CancelJob Cancel the job with the given jobID, Returns true if the job was found and cancelled.
func (jm *JobManager) CancelJob(jobId string) bool {
	job := Job{
		jobId:   jobId,
		jobType: CancelJob,
	}
	jm.jobs <- job
	return <-jm.cancelJobs
}

// RegisterJob regsiter a job in the JobManager context
func (jm *JobManager) RegisterJob(jobId string, t workerpool.Task, ctx context.CancelFunc, workPoolType *workerpool.TaskManager) {
	job := Job{
		jobId:        jobId,
		task:         t,
		ctx:          ctx,
		workPoolType: workPoolType,
		jobType:      RegisterJob,
	}
	jm.jobs <- job
}

// RemoveJob
func (jm *JobManager) RemoveJob(jobId string) {
	job := Job{
		jobId:   jobId,
		jobType: RemoveJob,
	}
	jm.jobs <- job
}

// start the job manager,a go routine which looks out for any incoming jobs.
func (jm *JobManager) Start() {

	go func() {
		for {
			select {
			case job, ok := <-jm.jobs:
				if !ok {
					return
				}
				jobType := job.jobType

				switch jobType {

				case RemoveJob:
					log.Info(log.Fields{"job_id": job.jobId}, "Removed Job")
					jm.jobLock.Lock()
					delete(jm.jobContext, job.jobId)
					jm.jobLock.Unlock()

				case RegisterJob:
					log.Info(log.Fields{"job_id": job.jobId}, "Registered Job")
					jm.jobLock.Lock()
					jm.jobContext[job.jobId] = job.ctx
					job.workPoolType.SubmitTask(job.task) //submit the job to a
					jm.jobLock.Unlock()

				case CancelJob:
					jm.jobLock.Lock()
					if cancel, ok := jm.jobContext[job.jobId]; ok {
						cancel()
						log.Info(log.Fields{"job_id": job.jobId}, "Cancelled job success")
						jm.cancelJobs <- true
					} else {
						log.Info(log.Fields{"job_id": job.jobId}, "Cancelled job failed")
						jm.cancelJobs <- false
					}
					jm.jobLock.Unlock()

				}

			case <-jm.quit:
				return
			}
		}
	}()
}

// not used...pause job needed?
func (jm *JobManager) Stop() {
	go func() {
		jm.quit <- true
		close(jm.quit)
		close(jm.jobs)
	}()
	log.Info(log.Fields{"name": "global job manager"}, "stopped job manager")
}

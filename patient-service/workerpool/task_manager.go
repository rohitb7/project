package workerpool

import (
	log "github.com/sirupsen/logrus"
)

type TaskManager struct {
	// A pool of workers channels that are registered with the dispatcher
	workerPool chan chan Task
	maxWorker  int
	taskQueue  chan Task
	quit       chan bool
	worker     []*Worker
	name       string
}

// Create your own worker pool as task manager
func NewTaskManager(maxWorkers int, name string, queueSize int) *TaskManager {
	pool := make(chan chan Task, maxWorkers)
	return &TaskManager{
		workerPool: pool,
		maxWorker:  maxWorkers,
		quit:       make(chan bool),
		taskQueue:  make(chan Task, queueSize),
		name:       name,
	}
}

// Creates new instance
// Start task managers with N workers
func (tm *TaskManager) Start() {
	// starting n number of workers
	for i := 0; i < tm.maxWorker; i++ {
		worker := NewWorker(tm.workerPool, i, tm.name)
		tm.worker = append(tm.worker, worker)
		worker.Start()
	}

	go tm.dispatch()
	log.Info(log.Fields{"name": tm.name}, "started task manager")
}

// Start allocating tasks to workers
func (tm *TaskManager) dispatch() {
	for {
		select {
		case task, open := <-tm.taskQueue: //nolint
			if !open {
				return
			}
			// a task request has been received
			//go func(task Task) {
			// try to obtain a worker task channel that is available.
			// this will block until a worker is idle
			taskChannel := <-tm.workerPool

			// dispatch the task to the worker task channel
			taskChannel <- task
			//}(task)

		case <-tm.quit:
			return
		}
	}
}

// Add runnable task to work queue. This call is not re-entrant. Do not call submitTask with async job from within another
// async job routine . It can lead to dead lock, if number of jobs > number of workers
func (tm *TaskManager) SubmitTask(t Task) {
	tm.taskQueue <- t
}

// Get Job Queue Occupancy as percentage
func (tm *TaskManager) IsQueueFull() bool {
	//lets keep some buffer
	return len(tm.taskQueue) >= (cap(tm.taskQueue) - 100)
}

// Stop all workers and task controller
func (tm *TaskManager) Stop() {
	for i, _ := range tm.worker { //nolint
		tm.worker[i].Stop()
	}
	go func() {
		tm.quit <- true
		close(tm.quit)
		close(tm.taskQueue)
	}()
	log.Info(log.Fields{"name": tm.name}, "stopped task manager")
}

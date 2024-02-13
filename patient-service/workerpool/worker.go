package workerpool

import (
	log "github.com/sirupsen/logrus"
)

// Task  with Runnable interface represents the generic work to be run
type Task interface {
	Run() error
}

// Worker represents the worker that executes the Task
type Worker struct {
	name        string
	id          int
	logCtx      log.Fields
	workerPool  chan chan Task
	taskChannel chan Task
	quit        chan bool
}

// Returns new worker
func NewWorker(workerPool chan chan Task, id int, name string) *Worker {
	return &Worker{
		name:        name,
		id:          id,
		logCtx:      log.Fields{"worker_id": id, "name": name},
		workerPool:  workerPool,
		taskChannel: make(chan Task),
		quit:        make(chan bool)}
}

// Start method starts the run loop for the worker, listening for a quit channel in
// case we need to stop it
func (w *Worker) Start() {
	go func() {
		for {
			// register the current worker into the worker queue.
			w.workerPool <- w.taskChannel

			select {
			case task, _ := <-w.taskChannel: //nolint
				// we have received a work request.
				_ = task.Run()

			case <-w.quit:
				// we have received a signal to stop
				return
			}
		}
	}()
	log.Info(w.logCtx, "started worker")
}

// Stop signals the worker to stop listening for work requests.
func (w *Worker) Stop() {
	go func() {
		w.quit <- true
		close(w.quit)
	}()
	log.Debug(w.logCtx, "stopped worker")
}

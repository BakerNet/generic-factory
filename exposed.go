package factory

import (
	"context"
)

// Job represents the data which will be processed by the workers
type Job interface {
	// Process is the function called by the worker goroutines.  Put your logic
	// in your data's Process method.
	// Note:  worker will also pre-process data with Register(ed) callbacks
	Process() error
}

// Factory represents a worker factory.  Workers process jobs requests
// concurrently.
type Factory interface {
	// Register callback to be called on each job received by a worker before
	// processing the Job - may Register an arbitrary number of callbacks
	Register(func(Job))

	// Dispatch job to an available worker.  Sends a ClosedFactoryError if job
	// is not completed before Factory has shut down.  Else sends error from
	// Job.Process
	Dispatch(Job) chan error

	// Close will stop all workers and prevent future dispatch jobs from being
	// handled.  Blocks until all worker goroutines have cleaned up
	Close()
}

// NewFactory returns a Factory - will fire up numWorkers worker routines.  Will
// close when ctx is done.
func NewFactory(ctx context.Context, numWorkers uint) Factory {
	f := &factory{
		state: state{
			callbacks: []func(Job){},
			jobCh:     make(chan job, numWorkers),
		},
		ctx:        ctx,
		quit:       make(chan struct{}),
		doneCh:     make(chan struct{}),
		dispatchCh: make(chan job),
		workers:    make([]*worker, numWorkers),
	}
	for i := uint(0); i < numWorkers; i++ {
		f.workers[i] = newWorker(f.state)
	}
	go f.manage()
	return f
}

// ClosedFactoryError - error returned by Dispatch channel if job was not
// handled due to the Factory being closed
type ClosedFactoryError struct{}

func (ClosedFactoryError) Error() string {
	return "factory closed before job could be handled"
}

package factory

import (
	"context"
)

// Factory represents a worker factory.  Workers process jobs requests concurrently.
type Factory interface {
	// Register callback to be called on each job received by a worker - may Register an arbitrary number of callbacks
	Register(func(interface{}))
	// Dispatch job to an available worker - returned channel will be closed when worker job has been processed.  Returns a ClosedFactoryError if job is not completed before Factory has shut down
	Dispatch(interface{}) chan error
	// Close will stop all workers and prevent future dispatch jobs from being handled
	Close()
}

// NewFactory returns a Factory - will fire up numWorkers worker routines.  Will close when ctx is done.
func NewFactory(ctx context.Context, numWorkers uint) Factory {
	f := &factory{
		state: state{
			callbacks: []func(interface{}){},
			jobCh:     make(chan job, numWorkers),
		},
		ctx:        ctx,
		done:       make(chan struct{}),
		dispatchCh: make(chan job),
		workers:    make([]*worker, numWorkers),
	}
	for i := uint(0); i < numWorkers; i++ {
		f.workers[i] = newWorker(f.state)
	}
	go f.manage()
	return f
}

// ClosedFactoryError - error returned by Dispatch channel if job was not handled due to the Factory being closed
type ClosedFactoryError struct{}

func (ClosedFactoryError) Error() string {
	return "Factory closed before job could be handled"
}
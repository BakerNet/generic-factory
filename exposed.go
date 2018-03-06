package factory

import (
	"context"
	"sync"
)

// Job represents the data which will be processed by the workers
type Job interface {
	// Process is the function called by the worker goroutines.  Put your logic
	// in your data's Process method.
	// Note:  worker will also pre-process data with Register(ed) callbacks
	Process() error
}

// NewFactory returns a Factory - will fire up numWorkers worker routines.  Will
// close when ctx is done.
func NewFactory(ctx context.Context, numWorkers uint) *Factory {
	f := &Factory{
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
	return f
}

// Factory represents a worker factory.  Workers process jobs requests
// concurrently.
type Factory struct {
	sync.Mutex
	state
	ctx      context.Context
	quit     chan struct{}
	isClosed bool
	// signal workers have all cleaned up
	doneCh     chan struct{}
	dispatchCh chan job
	workers    []*worker
}

// Run starts processing jobs for the factory - should be run in goroutine
func (f *Factory) Run() {
	go func() {
		select {
		case <-f.ctx.Done():
			f.Close()
		case <-f.quit:
		}
	}()
	for {
		select {
		case job := <-f.dispatchCh:
			f.jobCh <- job
		case <-f.quit:
			for _, w := range f.workers {
				w.Stop()
			}
			for _, w := range f.workers {
				<-w.doneCh
			}
			close(f.doneCh)
			return
		}
	}
}

// Register callback to be called on each job received by a worker before
// processing the Job - may Register an arbitrary number of callbacks
func (f *Factory) Register(cb func(Job)) {
	f.Lock()
	defer f.Unlock()
	f.callbacks = append(f.callbacks, cb)
	for _, w := range f.workers {
		w.Update(f.state)
	}
}

// Dispatch job to an available worker.  Sends a ClosedFactoryError if job
// is not completed before Factory has shut down.  Else sends error from
// Job.Process
func (f *Factory) Dispatch(data Job) chan error {
	dc := make(chan error, 1)
	go func(j job) {
		select {
		case f.dispatchCh <- j:
			return
		case <-f.quit:
			dc <- ClosedFactoryError{}
			return
		}
	}(job{data, dc})
	return dc
}

// Close will stop all workers and prevent future dispatch jobs from being
// handled.  Blocks until all worker goroutines have cleaned up
func (f *Factory) Close() {
	f.Lock()
	defer f.Unlock()
	if f.isClosed {
		return
	}
	f.isClosed = true
	close(f.quit)
	<-f.doneCh
}

// ClosedFactoryError - error returned by Dispatch channel if job was not
// handled due to the Factory being closed
type ClosedFactoryError struct{}

func (ClosedFactoryError) Error() string {
	return "factory closed before job could be handled"
}

package factory

import (
	"context"
	"sync"
)

type state struct {
	callbacks []func(Job)
	jobCh     chan job
}

type factory struct {
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

func (f *factory) Register(cb func(Job)) {
	f.Lock()
	defer f.Unlock()
	f.callbacks = append(f.callbacks, cb)
	for _, w := range f.workers {
		w.Update(f.state)
	}
}

func (f *factory) Dispatch(data Job) chan error {
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

func (f *factory) Close() {
	f.Lock()
	defer f.Unlock()
	if f.isClosed {
		return
	}
	f.isClosed = true
	close(f.quit)
	<-f.doneCh
}

func (f *factory) manage() {
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

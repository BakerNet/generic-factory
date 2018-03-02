package factory

import (
	"context"
	"sync"
)

type state struct {
	callbacks []func(interface{})
	jobCh     chan job
}

type factory struct {
	sync.Mutex
	state
	ctx        context.Context
	done       chan struct{}
	isClosed   bool
	dispatchCh chan job
	workers    []*worker
}

func (f *factory) Register(cb func(interface{})) {
	f.Lock()
	defer f.Unlock()
	f.callbacks = append(f.callbacks, cb)
	for _, w := range f.workers {
		w.Update(f.state)
	}
}

func (f *factory) Dispatch(data interface{}) chan error {
	dc := make(chan error, 1)
	go func(j job) {
		select {
		case f.dispatchCh <- j:
			return
		case <-f.done:
			dc <- ClosedFactoryError{}
			close(dc)
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
	close(f.done)
}

func (f *factory) manage() {
	go func() {
		select {
		case <-f.ctx.Done():
			f.Close()
		case <-f.done:
		}
	}()
	for {
		select {
		case job := <-f.dispatchCh:
			f.jobCh <- job
		case <-f.done:
			for _, w := range f.workers {
				w.Stop()
			}
			return
		}
	}
}

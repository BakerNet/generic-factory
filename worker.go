package factory

type worker struct {
	sent         uint64
	fs           state
	updateCh     chan struct{}
	updateDoneCh chan struct{}
	quit         chan struct{}
	done         bool
}

func newWorker(fs state) *worker {
	worker := &worker{
		fs:           fs,
		updateCh:     make(chan struct{}),
		updateDoneCh: make(chan struct{}),
		quit:         make(chan struct{}),
	}
	go worker.start()
	return worker
}

func (w *worker) start() {
	for {
		select {
		case job := <-w.fs.jobCh:
			for _, f := range w.fs.callbacks {
				f(job.data)
			}
			err := job.data.Process()
			job.doneCh <- err
		case <-w.updateCh:
			<-w.updateDoneCh
		case <-w.quit:
			select {
			case job := <-w.fs.jobCh:
				job.doneCh <- ClosedFactoryError{}
			default:
			}
			return
		}
	}
}

func (w *worker) Update(fs state) {
	if w.done {
		return
	}
	w.updateCh <- struct{}{}
	w.fs = fs
	w.updateDoneCh <- struct{}{}
}

func (w *worker) Stop() {
	if w.done {
		return
	}
	w.done = true
	close(w.quit)
}

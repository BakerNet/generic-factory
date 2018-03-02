package factory

import "testing"

func TestWorkerStoppedPreventsUpdateAndStop(t *testing.T) {
	w := &worker{
		updateCh:     make(chan struct{}),
		updateDoneCh: make(chan struct{}),
		quit:         make(chan struct{}),
	}
	go w.Update(state{})
	<-w.updateCh
	select {
	case <-w.updateDoneCh:
	default:
		t.Error("Expected to be able to read from updateDoneCh")
	}
	w.Stop()
	select {
	case <-w.quit:
	default:
		t.Error("Expected w.quit to be closed")
	}
	// w.Update would block if not exit early
	w.Update(state{})
	// Stop would panic on close if not exited early
	w.Stop()
}

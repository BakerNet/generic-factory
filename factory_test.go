package factory

import (
	"testing"
)

func TestRegister(t *testing.T) {
	f := &factory{}
	if len(f.callbacks) != 0 {
		t.Errorf("Expected no callbacks, got: %d", len(f.callbacks))
	}
	f.Register(func(interface{}) {})
	f.Register(func(interface{}) {})
	if len(f.callbacks) != 2 {
		t.Errorf("Expected 2 callbacks, got: %d", len(f.callbacks))
	}
}

func TestClosedFactoryErrrorFromDispatch(t *testing.T) {
	f := &factory{done: make(chan struct{})}
	dc := f.Dispatch(1)
	f.Close()
	if err := <-dc; err != nil {
		switch err.(type) {
		case ClosedFactoryError:
		default:
			t.Errorf("Expected ClosedFacotryError, got: %s", err)
		}
	} else {
		t.Errorf("Expected error from dc")
	}
}

func TestCloseMultipleTimes(t *testing.T) {
	f := &factory{done: make(chan struct{})}
	f.Close()
	select {
	case <-f.done:
	default:
		t.Error("Expected f.done to be closed")
	}
	// Close would panic if not exit early
	f.Close()
}
package factory

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func ExampleNewFactory_output() {
	ctx := context.Background()

	intFactory := NewFactory(ctx, 2)

	intFactory.Register(func(data interface{}) {
		fmt.Println(5 + data.(int))
	})

	doneChans := make([]chan error, 2)
	doneChans[0] = intFactory.Dispatch(2)
	doneChans[1] = intFactory.Dispatch(4)

	for _, dc := range doneChans {
		if err := <-dc; err != nil {
			fmt.Println("Error: ", err)
		}
	}

	intFactory.Close()
	// Unordered output: 7
	// 9
}

func TestFactoryErrorAfterClose(t *testing.T) {
	ctx := context.Background()

	f := NewFactory(ctx, 2)
	f.Close()
	dc := f.Dispatch(1)

	if err := <-dc; err == nil {
		t.Fatalf("Expected error for dispatch after close")
	}
}

func TestFactoryErrorAfterCtxDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	f := NewFactory(ctx, 2)
	cancel()
	time.Sleep(1 * time.Millisecond)
	dc := f.Dispatch(1)

	if err := <-dc; err == nil {
		t.Fatalf("Expected error for dispatch after close")
	}
}

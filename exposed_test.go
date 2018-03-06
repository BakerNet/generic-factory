package factory

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type intJob int

func (i *intJob) Process() error {
	fmt.Println(*i)
	return nil
}

func ExampleNewFactory_output() {
	/*
		type intJob int
		// Here, we use pointer to our data so we can register a function
		// which will modify the data before we call Process
		func (i *intJob) Process() error {
			fmt.Println(*i)
			return nil
		}
	*/

	ctx := context.Background()
	i := intJob(2)
	i2 := intJob(4)

	intFactory := NewFactory(ctx, 2)
	go intFactory.Run()
	intFactory.Register(func(j Job) {
		d := j.(*intJob)
		*d = *d + 5
	})

	doneChans := make([]chan error, 2)
	doneChans[0] = intFactory.Dispatch(&i)
	doneChans[1] = intFactory.Dispatch(&i2)

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
	i := intJob(1)

	f := NewFactory(ctx, 2)
	go f.Run()
	f.Close()
	dc := f.Dispatch(&i)

	if err := <-dc; err == nil {
		t.Fatalf("Expected error for dispatch after close")
	}
}

func TestFactoryErrorAfterCtxDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	i := intJob(1)

	f := NewFactory(ctx, 2)
	go f.Run()
	cancel()
	time.Sleep(1 * time.Millisecond)
	dc := f.Dispatch(&i)

	if err := <-dc; err == nil {
		t.Fatalf("Expected error for dispatch after close")
	}
}

func TestClosedFactoryErrorMsg(t *testing.T) {
	err := ClosedFactoryError{}
	if err.Error() != "factory closed before job could be handled" {
		t.Errorf("Wrong error messa")
	}
}

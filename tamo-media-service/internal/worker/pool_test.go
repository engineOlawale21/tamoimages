package worker

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestRunUsesBoundedConcurrency(t *testing.T) {
	jobs := make(chan int, 12)
	for job := range 12 {
		jobs <- job
	}
	close(jobs)
	var active int32
	var maximum int32
	err := Run(context.Background(), 3, jobs, func(context.Context, int) error {
		current := atomic.AddInt32(&active, 1)
		for current > atomic.LoadInt32(&maximum) {
			if atomic.CompareAndSwapInt32(&maximum, atomic.LoadInt32(&maximum), current) {
				break
			}
		}
		time.Sleep(time.Millisecond)
		atomic.AddInt32(&active, -1)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if maximum > 3 || maximum < 2 {
		t.Fatalf("unexpected maximum concurrency %d", maximum)
	}
}

func TestRunCancelsWorkersAfterHandlerFailure(t *testing.T) {
	jobs := make(chan int, 2)
	jobs <- 1
	jobs <- 2
	close(jobs)
	expected := errors.New("processing failed")
	err := Run(context.Background(), 1, jobs, func(context.Context, int) error { return expected })
	if !errors.Is(err, expected) {
		t.Fatalf("expected handler error, got %v", err)
	}
}

func TestRunRejectsInvalidConcurrency(t *testing.T) {
	if err := Run(context.Background(), 0, make(chan int), func(context.Context, int) error { return nil }); err == nil {
		t.Fatal("expected invalid concurrency error")
	}
}

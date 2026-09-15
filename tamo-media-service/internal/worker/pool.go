package worker

import (
	"context"
	"fmt"
	"sync"
)

type Handler[T any] func(context.Context, T) error

// Run drains jobs with bounded concurrency. Cancellation stops admission of new
// work and waits for admitted handlers to finish before returning.
func Run[T any](ctx context.Context, concurrency int, jobs <-chan T, handler Handler[T]) error {
	if concurrency < 1 {
		return fmt.Errorf("worker concurrency must be positive")
	}
	workerContext, cancel := context.WithCancel(ctx)
	defer cancel()
	var workers sync.WaitGroup
	var firstError error
	var errorOnce sync.Once
	workers.Add(concurrency)
	for range concurrency {
		go func() {
			defer workers.Done()
			for {
				select {
				case <-workerContext.Done():
					return
				case job, ok := <-jobs:
					if !ok {
						return
					}
					if err := handler(workerContext, job); err != nil {
						errorOnce.Do(func() {
							firstError = err
							cancel()
						})
						return
					}
				}
			}
		}()
	}
	workers.Wait()
	return firstError
}

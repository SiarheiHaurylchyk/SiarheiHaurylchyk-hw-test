package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if len(tasks) == 0 {
		return nil
	}

	taskChan := make(chan Task)
	var errCount atomic.Int32
	wg := sync.WaitGroup{}

	for i := 0; i < n; i++ {
		wg.Add(1)
		go worker(taskChan, &errCount, m, &wg)
	}

	for _, task := range tasks {
		if m > 0 && errCount.Load() >= int32(m) {
			break
		}
		taskChan <- task
	}

	close(taskChan)
	wg.Wait()

	if m > 0 && errCount.Load() >= int32(m) {
		return ErrErrorsLimitExceeded
	}
	return nil
}

func worker(taskChan <-chan Task, errCount *atomic.Int32, m int, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range taskChan {
		err := task()

		if err != nil && m > 0 {
			errCount.Add(1)
		}
	}
}

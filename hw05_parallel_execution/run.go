package hw05parallelexecution

import (
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if len(tasks) == 0 {
		return nil
	}
	if m < 0 {
		m = 0
	}
	if n <= 0 {
		n = 1
	}

	taskChan := make(chan Task)
	resultChan := make(chan error)
	wg := sync.WaitGroup{}

	for i := 0; i < n; i++ {
		wg.Add(1)
		go worker(taskChan, resultChan, &wg)
	}

	var errCount int
	idx, total := 0, len(tasks)

	for idx < total {
		batch := n
		left := total - idx
		if left < batch {
			batch = left
		}

		for i := 0; i < batch; i++ {
			taskChan <- tasks[idx]
			idx++
		}
		for i := 0; i < batch; i++ {
			if err := <-resultChan; err != nil {
				errCount++
			}
		}

		if m > 0 && errCount >= m {
			break
		}
	}

	close(taskChan)
	wg.Wait()

	if m > 0 && errCount >= m {
		return ErrErrorsLimitExceeded
	}
	return nil
}

func worker(taskChan <-chan Task, resultChan chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range taskChan {
		resultChan <- task()
	}
}

package grinfo

import (
	"context"
	"sync"
)

type Worker struct {
	workerNum  int
	bufferSize int
}

func NewWorker(workerNum, bufferSize int) *Worker {
	if workerNum < 1 {
		workerNum = 1
	}
	if bufferSize < 1 {
		bufferSize = 1
	}
	return &Worker{
		workerNum:  workerNum,
		bufferSize: bufferSize,
	}
}

type Result struct {
	Err error
	Log *Log
}

func (w *Worker) Start(ctx context.Context, inC <-chan string) <-chan *Result {
	var (
		resultC = make(chan *Result, w.bufferSize)
		wg      sync.WaitGroup
	)

	wg.Add(w.workerNum)
	for i := 0; i < w.workerNum; i++ {
		go func() {
			defer wg.Done()

			for repoDir := range inC {
				logger := NewLogger(NewGit(repoDir))
				r, err := logger.Get(ctx)
				resultC <- &Result{
					Err: err,
					Log: r,
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultC)
	}()

	return resultC
}

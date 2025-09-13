package grinfo

import (
	"context"
	"errors"
	"iter"
	"log/slog"

	"golang.org/x/sync/errgroup"
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

func (w *Worker) All(ctx context.Context, lines iter.Seq[string]) iter.Seq[*Result] {
	eg, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(ctx)

	var (
		inC     = make(chan string, w.bufferSize)
		resultC = make(chan *Result, w.bufferSize)
		worker  = func() error {
			for repoDir := range inC {
				if IsDone(ctx) {
					return ctx.Err()
				}
				sl := slog.With(slog.String("dir", repoDir))
				sl.Debug("start process")

				logger := NewLogger(NewGit(repoDir))
				r, err := logger.Get(ctx)
				if err != nil && errors.Is(err, context.Canceled) {
					sl.Debug("cancel process")
					return err
				}

				sl.Debug("end process")
				resultC <- &Result{
					Err: err,
					Log: r,
				}
			}

			return nil
		}
	)

	for range w.workerNum {
		eg.Go(worker)
	}

	go func() {
		defer cancel()

		for line := range lines {
			if IsDone(ctx) {
				break
			}
			inC <- line
		}
		close(inC)

		_ = eg.Wait()
		close(resultC)
	}()

	return func(yield func(*Result) bool) {
		defer cancel()

		for r := range resultC {
			if !yield(r) {
				return
			}
		}
	}
}

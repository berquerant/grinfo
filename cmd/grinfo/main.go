package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/berquerant/grinfo"
)

const usage = `grinfo - find latest commits in local repositories and the corresponding remote repositories

Usage:

  grinfo [flags]

Flags:`

func Usage() {
	fmt.Fprintln(os.Stderr, usage)
	flag.PrintDefaults()
}

func writeErr(format string, v ...any) {
	fmt.Fprintf(os.Stderr, format, v...)
}

func main() {
	var (
		workerNum  = flag.Int("worker", 4, "")
		bufferSize = flag.Int("buffer", 100, "")
	)

	flag.Usage = Usage
	flag.Parse()

	if *bufferSize < 1 {
		*bufferSize = 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var (
		inC     = make(chan string, *bufferSize)
		resultC = grinfo.NewWorker(*workerNum, *bufferSize).Start(ctx, inC)
		doneC   = make(chan struct{})
		sc      = bufio.NewScanner(os.Stdin)
	)

	go func() {
		defer close(doneC)

		for r := range resultC {
			if err := r.Err; err != nil {
				writeErr("%v\n", err)
				continue
			}
			b, err := json.Marshal(r.Log)
			if err != nil {
				writeErr("%v\n", err)
				continue
			}
			fmt.Printf("%s\n", b)
		}
	}()

	for sc.Scan() {
		inC <- sc.Text()
	}
	if err := sc.Err(); err != nil {
		writeErr("%v\n", err)
	}
	close(inC)
	<-doneC
}

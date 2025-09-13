package main

import (
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

Example:

  echo path/to/local/repo/dir | grinfo

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

	var lines grinfo.Lines
	for r := range grinfo.NewWorker(*workerNum, *bufferSize).
		All(ctx, lines.All(os.Stdin)) {
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

	if err := lines.Err(); err != nil {
		writeErr("%v\n", err)
	}
}

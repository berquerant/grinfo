package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
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

func main() {
	var (
		workerNum  = flag.Int("worker", 4, "")
		bufferSize = flag.Int("buffer", 100, "")
		debug      = flag.Bool("debug", false, "enable debug log")
	)

	flag.Usage = Usage
	flag.Parse()
	setupLogger(os.Stderr, *debug)

	if *bufferSize < 1 {
		*bufferSize = 1
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var lines grinfo.Lines
	for r := range grinfo.NewWorker(*workerNum, *bufferSize).
		All(ctx, lines.All(os.Stdin)) {
		if err := r.Err; err != nil {
			slog.Error("from result", slog.String("error", err.Error()))
			continue
		}
		b, err := json.Marshal(r.Log)
		if err != nil {
			slog.Error("failed to marshal log", slog.String("error", err.Error()))
			continue
		}
		fmt.Printf("%s\n", b)
	}

	if err := lines.Err(); err != nil {
		slog.Error("read lines", slog.String("error", err.Error()))
	}
}

func setupLogger(w io.Writer, debug bool) {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	handler := slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: level,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}

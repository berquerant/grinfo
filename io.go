package grinfo

import (
	"bufio"
	"io"
	"iter"
)

type Lines struct {
	err error
}

func (n Lines) Err() error {
	return n.err
}

func (n *Lines) All(r io.Reader) iter.Seq[string] {
	var (
		sc     = bufio.NewScanner(r)
		setErr = func(err error) {
			n.err = err
		}
	)
	return func(yield func(string) bool) {
		defer func() {
			setErr(sc.Err())
		}()

		for sc.Scan() {
			if !yield(sc.Text()) {
				return
			}
		}
	}
}

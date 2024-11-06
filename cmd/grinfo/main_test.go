package main_test

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestEndToEnd(t *testing.T) {
	e := newExecutor(t)
	defer e.close()

	if err := run(nil, e.cmd, "-h"); err != nil {
		t.Fatalf("%s help %v", e.cmd, err)
	}

	t.Run("grinfo self", func(t *testing.T) {
		pwd, err := os.Getwd()
		if err != nil {
			t.Error(err)
		}
		p, err := filepath.Abs(filepath.Join(pwd, "../../"))
		if err != nil {
			t.Error(err)
		}
		stdin := bytes.NewBufferString(p)
		if err := run(stdin, e.cmd); err != nil {
			t.Error(err)
		}
	})
}

func run(stdin io.Reader, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Dir = "."
	cmd.Stdin = stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type executor struct {
	dir string
	cmd string
}

func newExecutor(t *testing.T) *executor {
	t.Helper()
	e := &executor{}
	e.init(t)
	return e
}

func (e *executor) init(t *testing.T) {
	t.Helper()
	dir, err := os.MkdirTemp("", "grinfo")
	if err != nil {
		t.Fatal(err)
	}
	cmd := filepath.Join(dir, "grinfo")
	// build grinfo command
	if err := run(nil, "go", "build", "-o", cmd); err != nil {
		t.Fatal(err)
	}
	e.dir = dir
	e.cmd = cmd
}

func (e *executor) close() {
	os.RemoveAll(e.dir)
}

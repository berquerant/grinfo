package grinfo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func NewGit(dir string) *Git {
	return &Git{
		Dir: dir,
	}
}

type Git struct {
	Dir string
}

var (
	ErrGit = errors.New("Git")
)

type (
	GitLog struct {
		Hash      string
		Tree      string
		Parent    []string
		Author    GitLogMember
		Committer GitLogMember
		Message   string
	}

	GitLogMember struct {
		Name    string
		Email   string
		Date    time.Time
		RelDate string
	}
)

const (
	// https://git-scm.com/docs/git-log#_pretty_formats
	gitLogFormat     = "%H|%T|%P|%an|%ae|%ad|%ar|%cn|%ce|%cd|%cr|%s"
	gitLogDateFormat = "%Y-%m-%d %H:%M:%S"
)

func newGitLog(s string) (*GitLog, error) {
	var (
		n      = strings.Count(gitLogFormat, "|") + 1
		xs     = strings.SplitN(s, "|", n)
		genErr = func(err error) error {
			return fmt.Errorf("%w: invalid git log result: %s", errors.Join(ErrGit, err), s)
		}
	)
	if len(xs) != n {
		return nil, genErr(fmt.Errorf("want %d elements but got %d", n, len(xs)))
	}

	var (
		err       error
		author    GitLogMember
		committer GitLogMember
	)
	if author, err = newGitLogMember(xs[3], xs[4], xs[5], xs[6]); err != nil {
		return nil, genErr(err)
	}
	if committer, err = newGitLogMember(xs[7], xs[8], xs[9], xs[10]); err != nil {
		return nil, genErr(err)
	}

	return &GitLog{
		Hash:      xs[0],
		Tree:      xs[1],
		Parent:    strings.Split(xs[2], " "),
		Author:    author,
		Committer: committer,
		Message:   xs[11],
	}, nil
}

func parseGitLogTime(s string) (time.Time, error) {
	return time.Parse(time.DateTime, s)
}

func newGitLogMember(name, email, date, relDate string) (GitLogMember, error) {
	date = strings.Trim(date, "'")
	dateTime, err := parseGitLogTime(date)
	if err != nil {
		return GitLogMember{}, err
	}
	return GitLogMember{
		Name:    name,
		Email:   email,
		Date:    dateTime,
		RelDate: relDate,
	}, nil
}

func (g Git) Log(ctx context.Context, revision string) (*GitLog, error) {
	args := []string{
		"log",
		fmt.Sprintf("--pretty=format:'%s'", gitLogFormat),
		fmt.Sprintf("--date=format:'%s'", gitLogDateFormat),
		"-1",
		revision,
	}
	s, err := g.commandOutput(ctx, "git", args...)
	if err != nil {
		return nil, err
	}
	return newGitLog(strings.Trim(s, "'"))
}

func (g *Git) LatestRemoteCommitHash(ctx context.Context, url string) (string, error) {
	s, err := g.commandOutput(ctx, "git", "ls-remote", url, "HEAD")
	if err != nil {
		return "", err
	}
	xs := strings.SplitN(s, "\t", 2)
	if len(xs) != 2 || xs[0] == "" {
		return "", fmt.Errorf("%w: invalid ls-remote result: %s", ErrGit, s)
	}
	return xs[0], nil
}

func (g *Git) RemoteOriginURL(ctx context.Context) (string, error) {
	return g.commandOutput(ctx, "git", "config", "--get", "remote.origin.url")
}

func (g *Git) Fetch(ctx context.Context) error {
	return g.commandRun(ctx, "git", "fetch")
}

func (g *Git) LatestTag(ctx context.Context, revision string) (string, error) {
	return g.commandOutput(ctx, "git", "describe", "--abbrev=0", "--tags", revision)
}

func (g *Git) commandOutput(ctx context.Context, name string, args ...string) (string, error) {
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
		cmd    = g.command(ctx, name, args...)
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%w: %s", errors.Join(ErrGit, err), stderr)
	}
	return strings.TrimSpace(stdout.String()), nil
}

func (g *Git) commandRun(ctx context.Context, name string, args ...string) error {
	if err := g.command(ctx, name, args...).Run(); err != nil {
		return errors.Join(ErrGit, err)
	}
	return nil
}

func (g *Git) command(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = g.Dir
	return cmd
}

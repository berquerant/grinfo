package grinfo

import (
	"context"
	"time"
)

type (
	Log struct {
		URL      string      `json:"url"`
		Dir      string      `json:"dir"`
		TimeDiff LogTimeDiff `json:"timediff"`
		Local    LogElement  `json:"local"`
		Remote   LogElement  `json:"remote"`
	}

	LogTimeDiff struct {
		String string `json:"string"`
		Second int64  `json:"second"`
		Day    int    `json:"day"`
	}

	LogElement struct {
		Hash      string    `json:"hash"`
		Tree      string    `json:"tree"`
		Parent    []string  `json:"parent"`
		Message   string    `json:"message"`
		Author    LogMember `json:"author"`
		Committer LogMember `json:"committer"`
	}

	LogMember struct {
		Name      string `json:"name"`
		Email     string `json:"email"`
		RelDate   string `json:"reldate"`
		Date      string `json:"date"`
		Timestamp int64  `json:"timestamp"`
	}

	Logger struct {
		cmd *Git
	}
)

func NewLogger(cmd *Git) *Logger {
	return &Logger{
		cmd: cmd,
	}
}

func newLogTimeDiff(local, remote *GitLogMember) LogTimeDiff {
	d := remote.Date.Sub(local.Date)
	return LogTimeDiff{
		String: d.String(),
		Second: int64(d.Seconds()),
		Day:    int(d.Hours()) / 24,
	}
}

func newLogElement(l *GitLog) LogElement {
	return LogElement{
		Hash:      l.Hash,
		Tree:      l.Tree,
		Parent:    l.Parent,
		Message:   l.Message,
		Author:    newLogMember(&l.Author),
		Committer: newLogMember(&l.Committer),
	}
}

func newLogMember(m *GitLogMember) LogMember {
	return LogMember{
		Name:      m.Name,
		Email:     m.Email,
		RelDate:   m.RelDate,
		Date:      m.Date.Format(time.DateTime),
		Timestamp: m.Date.Unix(),
	}
}

func (l *Logger) Get(ctx context.Context) (*Log, error) {
	url, err := l.cmd.RemoteOriginURL(ctx)
	if err != nil {
		return nil, err
	}

	local, err := l.cmd.Log(ctx, "HEAD")
	if err != nil {
		return nil, err
	}

	if err := l.cmd.Fetch(ctx); err != nil {
		return nil, err
	}
	remoteHash, err := l.cmd.LatestRemoteCommitHash(ctx, url)
	if err != nil {
		return nil, err
	}
	remote, err := l.cmd.Log(ctx, remoteHash)
	if err != nil {
		return nil, err
	}

	return &Log{
		URL:      url,
		Dir:      l.cmd.Dir,
		TimeDiff: newLogTimeDiff(&local.Author, &remote.Author),
		Local:    newLogElement(local),
		Remote:   newLogElement(remote),
	}, nil
}

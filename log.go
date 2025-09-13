package grinfo

import (
	"context"
	"time"
)

type (
	Log struct {
		URL       string      `json:"url"`
		Dir       string      `json:"dir"`
		Local     LogElement  `json:"local"`
		Remote    LogElement  `json:"remote"`
		LocalTag  *LogElement `json:"local_tag,omitempty"`
		RemoteTag *LogElement `json:"remote_tag,omitempty"`
	}

	LogTimeDiff struct {
		String string `json:"string"`
		Second int64  `json:"second"`
		Day    int    `json:"day"`
	}

	LogElement struct {
		Hash              string      `json:"hash"`
		Tree              string      `json:"tree"`
		Parent            []string    `json:"parent"`
		Message           string      `json:"message"`
		Author            LogMember   `json:"author"`
		Committer         LogMember   `json:"committer"`
		TimeDiffFromLocal LogTimeDiff `json:"time_diff_from_local"`
		Tag               string      `json:"tag,omitempty"`
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

func newLogTimeDiff(dest, src *GitLogMember) LogTimeDiff {
	d := dest.Date.Sub(src.Date)
	return LogTimeDiff{
		String: d.String(),
		Second: int64(d.Seconds()),
		Day:    int(d.Hours()) / 24,
	}
}

func newLogElement(l, local *GitLog) LogElement {
	return LogElement{
		Hash:              l.Hash,
		Tree:              l.Tree,
		Parent:            l.Parent,
		Message:           l.Message,
		Author:            newLogMember(&l.Author),
		Committer:         newLogMember(&l.Committer),
		TimeDiffFromLocal: newLogTimeDiff(&l.Author, &local.Author),
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
	result := &Log{
		Dir: l.cmd.Dir,
	}

	url, err := l.cmd.RemoteOriginURL(ctx)
	if err != nil {
		return nil, err
	}
	result.URL = url

	local, err := l.cmd.Log(ctx, "HEAD")
	if err != nil {
		return nil, err
	}
	result.Local = newLogElement(local, local)

	if localTagName, err := l.cmd.LatestTag(ctx, local.Hash); err == nil {
		localTag, err := l.cmd.Log(ctx, localTagName)
		if err != nil {
			return nil, err
		}
		x := newLogElement(localTag, local)
		x.Tag = localTagName
		result.LocalTag = &x
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
	result.Remote = newLogElement(remote, local)

	if remoteTagName, err := l.cmd.LatestTag(ctx, remoteHash); err == nil {
		remoteTag, err := l.cmd.Log(ctx, remoteTagName)
		if err != nil {
			return nil, err
		}
		x := newLogElement(remoteTag, local)
		x.Tag = remoteTagName
		result.RemoteTag = &x
	}

	return result, nil
}

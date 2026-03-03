package grinfo

import (
	"context"
	"time"
)

type (
	Log struct {
		URL                     string      `json:"url"`
		Dir                     string      `json:"dir"`
		Local                   LogElement  `json:"local"`
		Remote                  LogElement  `json:"remote"`
		RemoteMinimumRelease    *LogElement `json:"remote_minimum_release"`
		LocalTag                *LogElement `json:"local_tag,omitempty"`
		RemoteTag               *LogElement `json:"remote_tag,omitempty"`
		RemoteTagMinimumRelease *LogElement `json:"remote_tag_minimum_release"`
		Diff                    LogDiff     `json:"diff"`
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
		TimeDiffToNow     LogTimeDiff `json:"time_diff_to_now"`
		Tag               string      `json:"tag,omitempty"`
	}

	LogMember struct {
		Name      string `json:"name"`
		Email     string `json:"email"`
		RelDate   string `json:"reldate"`
		Date      string `json:"date"`
		Timestamp int64  `json:"timestamp"`
	}

	LogDiff struct {
		Commit LogDiffList `json:"commit"`
		Tag    LogDiffList `json:"tag"`
	}

	LogDiffList struct {
		List  []string `json:"list"`
		Count int      `json:"count"`
	}

	Logger struct {
		cmd               *Git
		minimumReleaseAge time.Duration
	}
)

func NewLogger(cmd *Git, minimumReleaseAge time.Duration) *Logger {
	return &Logger{
		cmd:               cmd,
		minimumReleaseAge: minimumReleaseAge,
	}
}

func newLogDiffList(list []string) LogDiffList {
	return LogDiffList{
		List:  list,
		Count: len(list),
	}
}

func newLogDiff(commit, tag LogDiffList) LogDiff {
	return LogDiff{
		Commit: commit,
		Tag:    tag,
	}
}

func newLogTimeDiff(dest, src time.Time) LogTimeDiff {
	d := dest.Sub(src)
	return LogTimeDiff{
		String: d.String(),
		Second: int64(d.Seconds()),
		Day:    int(d.Hours()) / 24,
	}
}

func newLogElement(l, local *GitLog, now time.Time) LogElement {
	return LogElement{
		Hash:              l.Hash,
		Tree:              l.Tree,
		Parent:            l.Parent,
		Message:           l.Message,
		Author:            newLogMember(&l.Author),
		Committer:         newLogMember(&l.Committer),
		TimeDiffFromLocal: newLogTimeDiff(l.Author.Date, local.Author.Date),
		TimeDiffToNow:     newLogTimeDiff(now, l.Author.Date),
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
	now := time.Now()
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
	result.Local = newLogElement(local, local, now)

	if localTagName, err := l.cmd.LatestTag(ctx, local.Hash); err == nil {
		localTag, err := l.cmd.Log(ctx, localTagName)
		if err != nil {
			return nil, err
		}
		x := newLogElement(localTag, local, now)
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
	result.Remote = newLogElement(remote, local, now)

	if remoteTagName, err := l.cmd.LatestTag(ctx, remoteHash); err == nil {
		remoteTag, err := l.cmd.Log(ctx, remoteTagName)
		if err != nil {
			return nil, err
		}
		x := newLogElement(remoteTag, local, now)
		x.Tag = remoteTagName
		result.RemoteTag = &x
	}

	if commit, err := l.cmd.LatestCommitWithMinimumReleaseAge(ctx, l.minimumReleaseAge); err == nil {
		x, err := l.cmd.Log(ctx, commit)
		if err != nil {
			return nil, err
		}
		result.RemoteMinimumRelease = new(newLogElement(x, local, now))
	}

	if tag, err := l.cmd.LatestTagWithMinimumReleaseAge(ctx, l.minimumReleaseAge); err == nil {
		x, err := l.cmd.Log(ctx, tag)
		if err != nil {
			return nil, err
		}
		y := newLogElement(x, local, now)
		y.Tag = tag
		result.RemoteTagMinimumRelease = &y
	}

	commitDiff, err := l.cmd.ListCommitDiff(ctx, local.Hash)
	if err != nil {
		return nil, err
	}
	tagDiff, err := l.cmd.ListTagDiff(ctx, local.Hash)
	if err != nil {
		return nil, err
	}
	result.Diff = newLogDiff(newLogDiffList(commitDiff), newLogDiffList(tagDiff))

	return result, nil
}

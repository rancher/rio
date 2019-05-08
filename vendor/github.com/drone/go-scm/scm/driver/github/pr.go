// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/drone/go-scm/scm/driver/internal/null"
)

type pullService struct {
	*issueService
}

func (s *pullService) Find(ctx context.Context, repo string, number int) (*scm.PullRequest, *scm.Response, error) {
	path := fmt.Sprintf("repos/%s/pulls/%d", repo, number)
	out := new(pr)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertPullRequest(out), res, err
}

func (s *pullService) List(ctx context.Context, repo string, opts scm.PullRequestListOptions) ([]*scm.PullRequest, *scm.Response, error) {
	path := fmt.Sprintf("repos/%s/pulls?%s", repo, encodePullRequestListOptions(opts))
	out := []*pr{}
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	return convertPullRequestList(out), res, err
}

func (s *pullService) ListChanges(ctx context.Context, repo string, number int, opts scm.ListOptions) ([]*scm.Change, *scm.Response, error) {
	path := fmt.Sprintf("repos/%s/pulls/%d/files?%s", repo, number, encodeListOptions(opts))
	out := []*file{}
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	return convertChangeList(out), res, err
}

func (s *pullService) Merge(ctx context.Context, repo string, number int) (*scm.Response, error) {
	path := fmt.Sprintf("repos/%s/pulls/%d/merge", repo, number)
	res, err := s.client.do(ctx, "PUT", path, nil, nil)
	return res, err
}

func (s *pullService) Close(ctx context.Context, repo string, number int) (*scm.Response, error) {
	path := fmt.Sprintf("repos/%s/pulls/%d", repo, number)
	data := map[string]string{"state": "closed"}
	res, err := s.client.do(ctx, "PATCH", path, &data, nil)
	return res, err
}

type pr struct {
	Number  int    `json:"number"`
	State   string `json:"state"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	DiffURL string `json:"diff_url"`
	User    struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	} `json:"user"`
	Head struct {
		Ref  string `json:"ref"`
		Sha  string `json:"sha"`
		User struct {
			Login     string `json:"login"`
			AvatarURL string `json:"avatar_url"`
		}
		Repo struct {
			FullName string `json:"full_name"`
		} `json:"repo"`
	} `json:"head"`
	Base struct {
		Ref  string `json:"ref"`
		Sha  string `json:"sha"`
		User struct {
			Login     string `json:"login"`
			AvatarURL string `json:"avatar_url"`
		}
	} `json:"base"`
	MergedAt  null.String `json:"merged_at"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type file struct {
	Sha       string `json:"sha"`
	Filename  string `json:"filename"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Changes   int    `json:"changes"`
}

func convertPullRequestList(from []*pr) []*scm.PullRequest {
	to := []*scm.PullRequest{}
	for _, v := range from {
		to = append(to, convertPullRequest(v))
	}
	return to
}

func convertPullRequest(from *pr) *scm.PullRequest {
	return &scm.PullRequest{
		Number: from.Number,
		Title:  from.Title,
		Body:   from.Body,
		Sha:    from.Head.Sha,
		Ref:    fmt.Sprintf("refs/pull/%d/head", from.Number),
		Source: from.Head.Ref,
		Target: from.Base.Ref,
		Fork:   from.Head.Repo.FullName,
		Link:   from.DiffURL,
		Closed: from.State != "open",
		Merged: from.MergedAt.String != "",
		Author: scm.User{
			Login:  from.User.Login,
			Avatar: from.User.AvatarURL,
		},
		Created: from.CreatedAt,
		Updated: from.UpdatedAt,
	}
}

func convertChangeList(from []*file) []*scm.Change {
	to := []*scm.Change{}
	for _, v := range from {
		to = append(to, convertChange(v))
	}
	return to
}

func convertChange(from *file) *scm.Change {
	return &scm.Change{
		Path:    from.Filename,
		Added:   from.Status == "added",
		Deleted: from.Status == "deleted",
		Renamed: from.Status == "moved",
	}
}

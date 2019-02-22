// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scm

import (
	"context"
	"time"
)

type (
	// Issue represents an issue.
	Issue struct {
		Number  int
		Title   string
		Body    string
		Link    string
		Labels  []string
		Closed  bool
		Locked  bool
		Author  User
		Created time.Time
		Updated time.Time
	}

	// IssueInput provides the input fields required for
	// creating or updating an issue.
	IssueInput struct {
		Title string
		Body  string
	}

	// IssueListOptions provides options for querying a
	// list of repository issues.
	IssueListOptions struct {
		Page   int
		Size   int
		Open   bool
		Closed bool
	}

	// Comment represents a comment.
	Comment struct {
		ID      int
		Body    string
		Author  User
		Created time.Time
		Updated time.Time
	}

	// CommentInput provides the input fields required for
	// creating an issue comment.
	CommentInput struct {
		Body string
	}

	// IssueService provides access to issue resources.
	IssueService interface {
		// Find returns the issue by number.
		Find(context.Context, string, int) (*Issue, *Response, error)

		// FindComment returns the issue comment.
		FindComment(context.Context, string, int, int) (*Comment, *Response, error)

		// List returns the repository issue list.
		List(context.Context, string, IssueListOptions) ([]*Issue, *Response, error)

		// ListComments returns the issue comment list.
		ListComments(context.Context, string, int, ListOptions) ([]*Comment, *Response, error)

		// Create creates a new issue.
		Create(context.Context, string, *IssueInput) (*Issue, *Response, error)

		// CreateComment creates a new issue comment.
		CreateComment(context.Context, string, int, *CommentInput) (*Comment, *Response, error)

		// DeleteComment deletes an issue comment.
		DeleteComment(context.Context, string, int, int) (*Response, error)

		// Close closes an issue.
		Close(context.Context, string, int) (*Response, error)

		// Lock locks an issue discussion.
		Lock(context.Context, string, int) (*Response, error)

		// Unlock unlocks an issue discussion.
		Unlock(context.Context, string, int) (*Response, error)
	}
)

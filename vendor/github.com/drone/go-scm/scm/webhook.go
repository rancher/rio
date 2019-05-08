// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scm

import (
	"errors"
	"net/http"
)

var (
	// ErrSignatureInvalid is returned when the webhook
	// signature is invalid or cannot be calculated.
	ErrSignatureInvalid = errors.New("Invalid webhook signature")

	// ErrUnknownEvent is returned when the webhook event
	// is not recognized by the system.
	ErrUnknownEvent = errors.New("Unknown webhook event")
)

type (
	// Webhook defines a webhook for repository events.
	Webhook interface {
		Repository() Repository
	}

	// PushHook represents a push hook, eg push events.
	PushHook struct {
		Ref    string
		Repo   Repository
		Commit Commit
		Sender User
	}

	// BranchHook represents a branch or tag event,
	// eg create and delete github event types.
	BranchHook struct {
		Ref    Reference
		Repo   Repository
		Action Action
		Sender User
	}

	// TagHook represents a tag event, eg create and delete
	// github event types.
	TagHook struct {
		Ref    Reference
		Repo   Repository
		Action Action
		Sender User
	}

	// IssueHook represents an issue event, eg issues.
	IssueHook struct {
		Action Action
		Repo   Repository
		Issue  Issue
		Sender User
	}

	// IssueCommentHook represents an issue comment event,
	// eg issue_comment.
	IssueCommentHook struct {
		Action  Action
		Repo    Repository
		Issue   Issue
		Comment Comment
		Sender  User
	}

	// PullRequestHook represents an pull request event,
	// eg pull_request.
	PullRequestHook struct {
		Action      Action
		Repo        Repository
		PullRequest PullRequest
		Sender      User
	}

	// PullRequestCommentHook represents an pull request
	// comment event, eg pull_request_comment.
	PullRequestCommentHook struct {
		Action      Action
		Repo        Repository
		PullRequest PullRequest
		Comment     Comment
		Sender      User
	}

	// ReviewCommentHook represents a pull request review
	// comment, eg pull_request_review_comment.
	ReviewCommentHook struct {
		Action      Action
		Repo        Repository
		PullRequest PullRequest
		Review      Review
	}

	// SecretFunc provides the Webhook parser with the
	// secret key used to validate webhook authenticity.
	SecretFunc func(webhook Webhook) (string, error)

	// WebhookService provides abstract functions for
	// parsing and validating webhooks requests.
	WebhookService interface {
		// Parse returns the parsed the repository webhook payload.
		Parse(req *http.Request, fn SecretFunc) (Webhook, error)
	}
)

// Repository() defines the repository webhook and provides
// a convenient way to get the associated repository without
// having to cast the type.

func (h *PushHook) Repository() Repository               { return h.Repo }
func (h *BranchHook) Repository() Repository             { return h.Repo }
func (h *TagHook) Repository() Repository                { return h.Repo }
func (h *IssueHook) Repository() Repository              { return h.Repo }
func (h *IssueCommentHook) Repository() Repository       { return h.Repo }
func (h *PullRequestHook) Repository() Repository        { return h.Repo }
func (h *PullRequestCommentHook) Repository() Repository { return h.Repo }
func (h *ReviewCommentHook) Repository() Repository      { return h.Repo }

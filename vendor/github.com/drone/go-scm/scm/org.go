// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scm

import (
	"context"
)

type (
	// Organization represents an organization account.
	Organization struct {
		Name   string
		Avatar string
	}

	// OrganizationService provides access to organization resources.
	OrganizationService interface {
		// Find returns the organization by name.
		Find(context.Context, string) (*Organization, *Response, error)

		// List returns the user organization list.
		List(context.Context, ListOptions) ([]*Organization, *Response, error)
	}
)

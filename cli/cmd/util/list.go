package util

import "github.com/rancher/norman/types"

func DefaultListOpts() *types.ListOpts {
	return &types.ListOpts{
		Filters: map[string]interface{}{
			"sort":  "created",
			"order": "desc",
		},
	}
}

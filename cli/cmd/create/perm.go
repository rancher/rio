package create

import (
	"strings"

	"fmt"

	"github.com/rancher/rio/cli/pkg/kv"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

var (
	verbs = map[string]bool{}
)

func init() {
	for _, perm := range v1beta1.WriteVerbs {
		verbs[perm] = true
	}
}

func ParsePermissions(perms []string) ([]client.Permission, error) {
	var result []client.Permission
	for _, perm := range perms {
		p, err := parsePerm(perm)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}

	return result, nil
}

func assignAPIGroupResource(result *client.Permission, input string) {
	apiGroup, resource := kv.Split(input, "/")
	if resource == "" {
		result.APIGroup = ""
		result.Resource = apiGroup
	} else {
		result.APIGroup = apiGroup
		result.Resource = resource
	}
}

func assignVerbs(result *client.Permission, input string) {
	if input == "read" {
		result.Verbs = v1beta1.ReadVerbs
	} else if input == "write" {
		result.Verbs = v1beta1.WriteVerbs
	} else {
		for _, perm := range strings.Split(input, ",") {
			result.Verbs = append(result.Verbs, strings.TrimSpace(perm))
		}
	}
}

func parsePerm(perm string) (client.Permission, error) {
	var result client.Permission

	if strings.HasPrefix(perm, "role=") {
		result.Role = strings.TrimPrefix(perm, "role=")
		return result, nil
	}

	perm = strings.TrimPrefix(perm, "rule=")

	parts := strings.Fields(perm)

	if len(parts) == 1 {
		result.Verbs = v1beta1.ReadVerbs
		assignAPIGroupResource(&result, parts[0])
	} else {
		assignVerbs(&result, parts[0])
		assignAPIGroupResource(&result, parts[1])
		if len(parts) == 3 {
			result.Name = parts[2]
		}
	}

	if len(parts) > 3 {
		return result, fmt.Errorf("invalid format")
	}

	return result, nil
}

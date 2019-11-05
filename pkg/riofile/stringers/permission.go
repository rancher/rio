package stringers

import (
	"fmt"
	"sort"
	"strings"

	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
	"github.com/rancher/wrangler/pkg/slice"
)

var (
	verbs = map[string]bool{}
)

var (
	ReadVerbs = []string{
		"get",
		"list",
		"watch",
	}
	WriteVerbs = []string{
		"create",
		"delete",
		"get",
		"list",
		"patch",
		"update",
		"watch",
	}
)

func init() {
	for _, perm := range WriteVerbs {
		verbs[perm] = true
	}
}

type PermissionStringer struct {
	v1.Permission
}

func (p PermissionStringer) MaybeString() interface{} {
	if p.Role != "" {
		return fmt.Sprintf("role=%s", p.Role)
	}

	sort.Strings(p.Verbs)

	buf := strings.Builder{}
	if slice.StringsEqual(WriteVerbs, p.Verbs) {
		buf.WriteString("write ")
	} else if len(p.Verbs) > 0 && !slice.StringsEqual(ReadVerbs, p.Verbs) {
		buf.WriteString(strings.Join(p.Verbs, ","))
		buf.WriteString(" ")
	}

	groups := p.APIGroup
	resources := p.Resource
	names := p.ResourceName

	if groups != "" || strings.Contains(resources, "/") || resources == "*" {
		buf.WriteString(groups)
		buf.WriteString("/")
	}

	buf.WriteString(resources)

	if names != "" {
		buf.WriteString(" ")
		buf.WriteString(names)
	}

	if len(p.URL) > 0 {
		buf.WriteString(" ")
		buf.WriteString("url=")
		buf.WriteString(p.URL)
	}

	return buf.String()
}

func ParsePermission(perm string) (result v1.Permission, err error) {
	if strings.HasPrefix(perm, "role=") {
		perm = strings.TrimSpace(strings.TrimPrefix(perm, "role="))
		return v1.Permission{
			Role: perm,
		}, nil
	}
	perm = strings.TrimSpace(strings.TrimPrefix(perm, "rule="))
	if perm == "" {
		return result, fmt.Errorf("empty rule found")
	}
	return parsePerm(perm)
}

func ParsePermissions(perms ...string) (result []v1.Permission, err error) {
	for _, perm := range perms {
		p, err := ParsePermission(perm)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return
}

func assignAPIGroupResource(result *v1.Permission, input string) {
	apiGroup, resource := kv.Split(input, "/")
	if resource == "" {
		result.Resource = apiGroup
	} else {
		result.APIGroup = apiGroup
		result.Resource = resource
	}
}

func assignVerbs(result *v1.Permission, input string) {
	switch input {
	case "read":
		result.Verbs = ReadVerbs
	case "write":
		result.Verbs = WriteVerbs
	default:
		for _, perm := range strings.Split(input, ",") {
			result.Verbs = append(result.Verbs, strings.TrimSpace(perm))
		}
	}
}

func parsePerm(perm string) (v1.Permission, error) {
	var result v1.Permission

	parts := filterURL(strings.Fields(perm), &result)

	if len(parts) == 1 {
		result.Verbs = ReadVerbs
		assignAPIGroupResource(&result, parts[0])
	} else {
		assignVerbs(&result, parts[0])
		assignAPIGroupResource(&result, parts[1])
		if len(parts) == 3 {
			result.ResourceName = parts[2]
		}
	}

	if len(parts) > 3 {
		return result, fmt.Errorf("invalid format")
	}

	return result, nil
}

func filterURL(parts []string, policy *v1.Permission) []string {
	var result []string
	for _, input := range parts {
		if strings.HasPrefix(input, "url=") {
			policy.URL = strings.TrimPrefix(input, "url=")
			continue
		}
		result = append(result, input)
	}
	return result
}

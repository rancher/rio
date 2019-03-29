package objectmappers

import (
	"fmt"
	"strings"

	"github.com/rancher/mapper/mappers"
	"github.com/rancher/mapper/slice"
	v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/kv"
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

func NewPermission(field string) mappers.ObjectsToSlice {
	return mappers.ObjectsToSlice{
		Field: field,
		NewObject: func() mappers.MaybeStringer {
			return &PermissionStringer{}
		},
		ToObject: func(obj interface{}) (interface{}, error) {
			str, ok := obj.(string)
			if !ok {
				return obj, nil
			}
			objs, err := ParsePermissions(str)
			if err != nil {
				return nil, err
			}
			return objs[0], nil
		},
	}
}

type PermissionStringer struct {
	v1.Permission
}

func (p PermissionStringer) MaybeString() interface{} {
	if p.Role != "" {
		return "role=" + p.Role
	}

	buf := strings.Builder{}
	if slice.StringsEqual(WriteVerbs, p.Verbs) {
		buf.WriteString("write ")
	} else if len(p.Verbs) > 0 && !slice.StringsEqual(ReadVerbs, p.Verbs) {
		buf.WriteString(strings.Join(p.Verbs, ","))
		buf.WriteString(" ")
	}

	if p.URL == "" {
		if p.APIGroup != "" || strings.Contains(p.Resource, "/") {
			buf.WriteString(p.APIGroup)
			buf.WriteString("/")
		}

		buf.WriteString(p.Resource)

		if p.Name != "" {
			buf.WriteString(" ")
			buf.WriteString(p.Name)
		}
	} else {
		buf.WriteString("url=")
		buf.WriteString(p.URL)
	}

	return buf.String()
}

func ParsePermissions(perms ...string) ([]v1.Permission, error) {
	var result []v1.Permission
	for _, perm := range perms {
		p, err := parsePerm(perm)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}

	return result, nil
}

func assignAPIGroupResource(result *v1.Permission, input string) {
	if strings.HasPrefix(input, "url=") {
		result.URL = strings.TrimPrefix(input, "url=")
		return
	}

	apiGroup, resource := kv.Split(input, "/")
	if resource == "" {
		result.APIGroup = ""
		result.Resource = apiGroup
	} else {
		result.APIGroup = apiGroup
		result.Resource = resource
	}
}

func assignVerbs(result *v1.Permission, input string) {
	if input == "read" {
		result.Verbs = ReadVerbs
	} else if input == "write" {
		result.Verbs = WriteVerbs
	} else {
		for _, perm := range strings.Split(input, ",") {
			result.Verbs = append(result.Verbs, strings.TrimSpace(perm))
		}
	}
}

func parsePerm(perm string) (v1.Permission, error) {
	var result v1.Permission

	if strings.HasPrefix(perm, "role=") {
		result.Role = strings.TrimPrefix(perm, "role=")
		return result, nil
	}

	perm = strings.TrimPrefix(perm, "rule=")

	parts := strings.Fields(perm)

	if len(parts) == 1 {
		result.Verbs = ReadVerbs
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

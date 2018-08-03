package v1beta1

import (
	"strings"

	"github.com/rancher/norman/types/slice"
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

type Permission struct {
	Role     string   `json:"role,omitempty"`
	Verbs    []string `json:"verbs,omitempty"`
	APIGroup string   `json:"apiGroup,omitempty"`
	Resource string   `json:"resource,omitempty"`
	Name     string   `json:"name,omitempty"`
}

func (p Permission) MaybeString() interface{} {
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

	if p.APIGroup != "" {
		buf.WriteString(p.APIGroup)
		buf.WriteString("/")
	}

	buf.WriteString(p.Resource)

	if p.Name != "" {
		buf.WriteString(" ")
		buf.WriteString(p.Name)
	}

	return buf.String()
}

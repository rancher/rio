package stringers

import (
	"fmt"
	"strings"

	"github.com/rancher/mapper/slice"
	"github.com/rancher/wrangler/pkg/kv"
	rbacv1 "k8s.io/api/rbac/v1"
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

type PolicyRuleStringer struct {
	rbacv1.PolicyRule
}

func (p PolicyRuleStringer) String() string {
	buf := strings.Builder{}
	if slice.StringsEqual(WriteVerbs, p.Verbs) {
		buf.WriteString("write ")
	} else if len(p.Verbs) > 0 && !slice.StringsEqual(ReadVerbs, p.Verbs) {
		buf.WriteString(strings.Join(p.Verbs, ","))
		buf.WriteString(" ")
	}

	groups := strings.Join(p.APIGroups, ",")
	resources := strings.Join(p.Resources, ",")
	names := strings.Join(p.ResourceNames, ",")

	if groups != "" || strings.Contains(resources, "/") {
		buf.WriteString(groups)
	}

	buf.WriteString(resources)

	if names != "" {
		buf.WriteString(" ")
		buf.WriteString(names)
	}

	if len(p.NonResourceURLs) > 0 {
		buf.WriteString("url=")
		buf.WriteString(strings.Join(p.NonResourceURLs, ","))
	}

	return buf.String()
}

func ParseRoles(perms ...string) []string {
	var result []string
	for _, perm := range perms {
		if strings.HasPrefix(perm, "role=") {
			perm = strings.TrimSpace(strings.TrimPrefix(perm, "role="))
			if perm != "" {
				result = append(result, perm)
			}
		}
	}

	return result
}

func ParsePolicyRules(perms ...string) ([]rbacv1.PolicyRule, error) {
	var result []rbacv1.PolicyRule
	for _, perm := range perms {
		if strings.HasPrefix(perm, "role=") {
			continue
		}
		perm = strings.TrimSpace(strings.TrimPrefix(perm, "rule="))
		if perm == "" {
			continue
		}
		p, err := parsePerm(perm)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}

	return result, nil
}

func assignAPIGroupResource(result *rbacv1.PolicyRule, input string) {
	apiGroup, resource := kv.Split(input, "/")
	if resource == "" {
		result.APIGroups = []string{""}
		result.Resources = strings.Split(apiGroup, ",")
	} else {
		result.APIGroups = strings.Split(apiGroup, ",")
		result.Resources = strings.Split(resource, ",")
	}
}

func assignVerbs(result *rbacv1.PolicyRule, input string) {
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

func parsePerm(perm string) (rbacv1.PolicyRule, error) {
	var result rbacv1.PolicyRule

	parts := filterURL(strings.Fields(perm), &result)

	if len(parts) == 1 {
		result.Verbs = ReadVerbs
		assignAPIGroupResource(&result, parts[0])
	} else {
		assignVerbs(&result, parts[0])
		assignAPIGroupResource(&result, parts[1])
		if len(parts) == 3 {
			result.ResourceNames = strings.Split(parts[2], ",")
		}
	}

	if len(parts) > 3 {
		return result, fmt.Errorf("invalid format")
	}

	return result, nil
}

func filterURL(parts []string, policy *rbacv1.PolicyRule) []string {
	var result []string
	for _, input := range parts {
		if strings.HasPrefix(input, "url=") {
			policy.NonResourceURLs = strings.Split(strings.TrimPrefix(input, "url="), ",")
			continue
		}
		result = append(result, input)
	}
	return result
}

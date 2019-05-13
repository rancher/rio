package stringers

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
)

var (
	pullNames = map[string]v1.PullPolicy{
		"always":         v1.PullAlways,
		"never":          v1.PullNever,
		"not-present":    v1.PullIfNotPresent,
		"if-not-present": v1.PullIfNotPresent,
		"ifnotpresent":   v1.PullIfNotPresent,
		"notpresent":     v1.PullIfNotPresent,
		"":               "",
	}
)

func ParseImagePullPolicy(policy string) (v1.PullPolicy, error) {
	ret, ok := pullNames[strings.ToLower(policy)]
	if !ok {
		return ret, fmt.Errorf("%s is not a valid image pull policy, must be Always, Never, or IfNotPresent", policy)
	}
	return ret, nil
}

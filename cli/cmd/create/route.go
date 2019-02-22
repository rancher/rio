package create

import (
	"strings"

	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func ParseStringMatch(str string) *riov1.StringMatch {
	if strings.HasSuffix(str, "*") {
		return &riov1.StringMatch{
			Prefix: str[:len(str)-1],
		}
	} else if IsRegexp(str) {
		return &riov1.StringMatch{
			Regexp: strings.TrimSuffix(strings.SplitN(str, "(", 2)[1], ")"),
		}
	}

	return &riov1.StringMatch{
		Exact: str,
	}
}

func IsRegexp(str string) bool {
	return (strings.HasPrefix(str, "regex(") || strings.HasPrefix(str, "regexp(")) && strings.HasSuffix(str, ")")
}

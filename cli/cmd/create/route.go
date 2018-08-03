package create

import (
	"strings"

	"github.com/rancher/rio/types/client/rio/v1beta1"
)

func ParseStringMatch(str string) *client.StringMatch {
	if strings.HasSuffix(str, "*") {
		return &client.StringMatch{
			Prefix: str[:len(str)-1],
		}
	} else if IsRegexp(str) {
		return &client.StringMatch{
			Regexp: strings.TrimSuffix(strings.SplitN(str, "(", 2)[1], ")"),
		}
	}

	return &client.StringMatch{
		Exact: str,
	}
}

func IsRegexp(str string) bool {
	return (strings.HasPrefix(str, "regex(") || strings.HasPrefix(str, "regexp(")) && strings.HasSuffix(str, ")")
}

package parse

import (
	"net/url"
	"strings"
)

func ParseTargetURL(target string) (*url.URL, error) {
	if !strings.HasPrefix(target, "https://") && !strings.HasPrefix(target, "http://") {
		target = "http://" + target
	}
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	return u, nil
}

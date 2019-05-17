package name

import (
	"strings"

	"github.com/rancher/wrangler/pkg/name"
)

func PublicDomain(s string) string {
	return name.Limit(strings.Replace(s, ".", "-", -1), 15)
}

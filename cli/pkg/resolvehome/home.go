package resolvehome

import (
	"os/user"
	"strings"

	"github.com/pkg/errors"
)

var (
	homes = []string{"$HOME", "${HOME}", "~"}
)

func Resolve(s string) (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", errors.Wrap(err, "determining current user")
	}

	for _, home := range homes {
		s = strings.Replace(s, home, u.HomeDir, -1)
	}

	return s, nil
}

package paths

import (
	"github.com/rancher/rio/cli/pkg/resolvehome"
	"github.com/urfave/cli"
)

const ()

func ConfigHome(app *cli.Context) (string, error) {
	return resolvehome.Resolve(confHome)
}

func Paths() (string, string, error) {
	ch, err := ConfigHome()
	if err != nil {
		return "", "", err
	}

	rio, err := RioConfPath()
	if err != nil {
		return "", "", err
	}

	return ch, rio, nil
}

// +build !k3s

package agent

import (
	"fmt"

	"github.com/urfave/cli"
)

func (a *Agent) Run(app *cli.Context) error {
	return fmt.Errorf("agent support is not compiled in, add \"-tags k3s\" to \"go build\"")
}

func RunAgent(server, token, dataDir, logFile string) error {
	return fmt.Errorf("agent support is not compiled in, add \"-tags k3s\" to \"go build\"")
}

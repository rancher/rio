// +build !k3s

package server

import (
	"fmt"

	"github.com/urfave/cli"
)

func (s *Server) Run(app *cli.Context) error {
	return fmt.Errorf("server support is not compiled in, add \"-tags k3s\" to \"go build\"")
}

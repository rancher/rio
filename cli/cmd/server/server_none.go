// +build !k8s

package server

import (
	"fmt"

	"github.com/urfave/cli"
)

func (s *Server) Run(app *cli.Context) error {
	return fmt.Errorf("server support is not compiled in, add \"-tags k8s\" to \"go build\"")
}

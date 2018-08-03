// +build !ctr

package ctr

import (
	"fmt"

	"github.com/urfave/cli"
)

func ctr(app *cli.Context) error {
	return fmt.Errorf("ctr not compiled in, add \"-tags ctr\" to build")
}

package project

import (
	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/pkg/clicontext"
)

type Add struct{}

func (a *Add) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return errors.New("at least one argument is need. Example: `rio project add PROJECT_NAME`")
	}
	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}
	var lastErr error
	for _, arg := range ctx.CLI.Args() {
		_, err := cluster.CreateProject(arg)
		if err != nil {
			lastErr = err
		}
	}
	if lastErr != nil {
		return lastErr
	}
	return nil
}

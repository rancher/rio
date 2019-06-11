package attach

import (
	"fmt"
	"time"

	"github.com/rancher/rio/cli/cmd/ps"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/tables"
)

type Attach struct {
	Timeout string `desc:"Timeout waiting for the container to be created to attach to" default:"1m"`
}

func (a *Attach) Run(ctx *clicontext.CLIContext) error {
	args := ctx.CLI.Args()
	if len(args) < 1 {
		return fmt.Errorf("at least one argument is required: CONTAINER")
	}

	timeout, err := time.ParseDuration(a.Timeout)
	if err != nil {
		return err
	}

	return RunAttach(ctx, timeout, ctx.CLI.Args()[0])
}

func RunAttach(ctx *clicontext.CLIContext, timeout time.Duration, container string) error {
	var pd *tables.PodData
	var err error

	deadline := time.Now().Add(timeout)
	for {
		pd, err = ps.ListFirstPod(ctx, true, container)
		if err != nil {
			return err
		}

		if (pd == nil || len(pd.Containers) == 0) && time.Now().Before(deadline) {
			time.Sleep(750 * time.Millisecond)
			continue
		}

		break
	}

	if pd == nil {
		return fmt.Errorf("failed to find a container for %s", container)
	}

	execArgs := []string{
		fmt.Sprintf("--pod-running-timeout=%s", timeout),
		pd.Pod.Name,
		"-c", pd.Containers[0].Name,
	}

	if pd.Containers[0].Stdin {
		execArgs = append(execArgs, "-i")
	}
	if pd.Containers[0].TTY {
		execArgs = append(execArgs, "-t")
	}

	return ctx.Kubectl(pd.Pod.Namespace, "attach", execArgs...)
}

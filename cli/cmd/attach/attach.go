package attach

import (
	"fmt"
	"time"

	"github.com/rancher/rio/cli/cmd/ps"
	"github.com/rancher/rio/cli/pkg/clicontext"
)

type Attach struct {
	I_Stdin bool   `desc:"Pass stdin to the container"`
	T_Tty   bool   `desc:"Stdin is a TTY"`
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

	return RunAttach(ctx, timeout, a.I_Stdin, a.T_Tty, ctx.CLI.Args()[0])
}

func RunAttach(ctx *clicontext.CLIContext, timeout time.Duration, stdin, tty bool, container string) error {
	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}

	var pd *ps.PodData

	deadline := time.Now().Add(timeout)
	for {
		pd, err = ps.ListFirstPod(ctx, true, container)
		if err != nil {
			return err
		}

		if (pd == nil || len(pd.Containers) == 0 || pd.Containers[0].State != "running") && time.Now().Before(deadline) {
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
	if stdin {
		execArgs = append(execArgs, "-i")
	}
	if tty {
		execArgs = append(execArgs, "-t")
	}

	return cluster.Kubectl(pd.Pod.Namespace, "attach", execArgs...)
}

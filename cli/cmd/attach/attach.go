package attach

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/rancher/rio/cli/cmd/util"

	"github.com/rancher/rio/cli/pkg/clicontext"
)

type Attach struct {
	Pod     string `desc:"Specify pod name, default to the first pod"`
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

	return RunAttach(ctx, timeout, ctx.CLI.Args()[0], a.Pod)
}

func RunAttach(ctx *clicontext.CLIContext, timeout time.Duration, service, podName string) error {
	var pod v1.Pod
	deadline := time.Now().Add(timeout)
	for {
		pods, err := util.ListPods(ctx, service)
		if err != nil {
			return err
		}

		if len(pods) == 0 {
			continue
		}

		if podName != "" {
			for _, p := range pods {
				if p.Name == podName {
					pod = p
					break
				}
			}
			if pod.Name == "" {
				continue
			}
		} else {
			pod = pods[0]
		}

		if time.Now().Before(deadline) {
			time.Sleep(750 * time.Millisecond)
			continue
		}

		break
	}

	if pod.Name == "" {
		return fmt.Errorf("failed to find pod for %s", service)
	}

	execArgs := []string{
		fmt.Sprintf("--pod-running-timeout=%s", timeout),
		pod.Name,
		"-c", pod.Spec.Containers[0].Name,
	}

	if pod.Spec.Containers[0].Stdin {
		execArgs = append(execArgs, "-i")
	}
	if pod.Spec.Containers[0].TTY {
		execArgs = append(execArgs, "-t")
	}

	return ctx.Kubectl(pod.Namespace, "attach", execArgs...)
}

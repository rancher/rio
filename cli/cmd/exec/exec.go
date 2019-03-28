package exec

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/cli/cmd/ps"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

type Exec struct {
	I_Stdin     bool   `desc:"Pass stdin to the container"`
	T_Tty       bool   `desc:"Stdin is a TTY"`
	C_Container string `desc:"Specific container in pod, default is first container"`
}

func (e *Exec) Run(ctx *clicontext.CLIContext) error {
	args := ctx.CLI.Args()
	if len(args) < 2 {
		return fmt.Errorf("at least two arguments are required CONTAINER CMD")
	}

	pd, err := ps.ListFirstPod(ctx, true, args[0])
	if err != nil {
		return err
	}

	if pd == nil {
		return fmt.Errorf("failed to find pod for %s, container \"%s\"", args[0], e.C_Container)
	}

	if e.C_Container == "" {
		parts := strings.SplitN(args[0], "/", 4)
		if len(parts) == 4 {
			e.C_Container = parts[3]
		}
	}

	container := findContainer(pd, e.C_Container)
	podNS, podName, containerName := pd.Pod.Namespace, pd.Pod.Name, container.Name

	execArgs := []string{}
	if logrus.GetLevel() >= logrus.DebugLevel {
		execArgs = append(execArgs, "-v=9")
	}
	if e.I_Stdin {
		execArgs = append(execArgs, "-i")
	}
	if e.T_Tty {
		execArgs = append(execArgs, "-t")
	}

	execArgs = append(execArgs, podName, "-c", containerName, "--")
	execArgs = append(execArgs, args[1:]...)

	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}
	return cluster.Kubectl(podNS, "exec", execArgs...)
}

func findContainer(pd *ps.PodData, name string) *v1.Container {
	for _, c := range pd.Containers {
		if c.Name == name {
			return &c
		}
	}

	return &pd.Containers[0]
}

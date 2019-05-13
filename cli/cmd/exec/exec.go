package exec

import (
	"fmt"
	"strings"

	"github.com/rancher/rio/cli/cmd/ps"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/tables"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

type Exec struct {
	I_Stdin     bool   `desc:"Pass stdin to the container"`
	T_Tty       bool   `desc:"Stdin is a TTY"`
	C_Container string `desc:"Specific container in pod, default is first container"`
	R_Revision  string `desc:"Specific revision , default is first revision"`
}

func (e *Exec) Run(ctx *clicontext.CLIContext) error {
	args := ctx.CLI.Args()
	if len(args) < 2 {
		return fmt.Errorf("at least two arguments are required CONTAINER CMD")
	}

	pds, err := ps.ListPods(ctx, true, args[0])
	if err != nil {
		return err
	}

	if len(pds) == 0 {
		return fmt.Errorf("failed to find pod for %s, container \"%s\"", args[0], e.C_Container)
	}

	pd := pds[0]
	if e.R_Revision != "" {
		for _, p := range pds {
			if p.Service.Version == e.R_Revision {
				pd = p
				break
			}
		}
	}

	if e.C_Container == "" {
		parts := strings.SplitN(args[0], "/", 4)
		if len(parts) == 4 {
			e.C_Container = parts[3]
		}
	}

	container := findContainer(&pd, e.C_Container)
	podNS, podName, containerName := pd.Pod.Namespace, pd.Pod.Name, container.Name

	var execArgs []string
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

	return ctx.Kubectl(podNS, "exec", execArgs...)
}

func findContainer(pd *tables.PodData, name string) *v1.Container {
	for _, c := range pd.Containers {
		if c.Name == name {
			return &c
		}
	}

	return &pd.Containers[0]
}

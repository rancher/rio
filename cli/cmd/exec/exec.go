package exec

import (
	"fmt"

	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

type Exec struct {
	C_Container string `desc:"Specify container in pod, default is first container"`
	Pod         string `desc:"Specify pod, default is first pod found"`
	I_Stdin     bool   `desc:"Pass stdin to the container"`
	T_Tty       bool   `desc:"Stdin is a TTY"`
}

func (e *Exec) Run(ctx *clicontext.CLIContext) error {
	args := ctx.CLI.Args()
	if len(args) < 2 {
		return fmt.Errorf("at least two arguments are required CONTAINER CMD")
	}

	pds, err := util.ListPods(ctx, args[0])
	if err != nil {
		return err
	}

	if len(pds) == 0 {
		return fmt.Errorf("failed to find any pod for service %s, container \"%s\"", args[0], e.C_Container)
	}

	var pod v1.Pod
	if e.Pod != "" {
		for _, p := range pds {
			if p.Name == e.Pod {
				pod = p
				break
			}
		}
		if pod.Name == "" {
			return fmt.Errorf("failed to find pod for service %s, pod name \"%s\"", args[0], e.Pod)
		}
	} else {
		pod = pds[0]
	}

	var con v1.Container
	if e.C_Container != "" {
		for _, c := range pod.Spec.Containers {
			if c.Name == e.C_Container {
				con = c
				break
			}
		}
		if con.Name == "" {
			return fmt.Errorf("failed to find pod for service %s, container name \"%s\"", args[0], e.C_Container)
		}
	} else {
		con = pod.Spec.Containers[0]
	}

	podNS, podName, containerName := pod.Namespace, pod.Name, con.Name

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

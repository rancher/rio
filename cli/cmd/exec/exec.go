package exec

import (
	"fmt"
	"os"

	"github.com/docker/docker/pkg/reexec"
	"github.com/rancher/rio/cli/cmd/ps"
	"github.com/rancher/rio/cli/server"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type Exec struct {
	I_Stdin     bool   `desc:"Pass stdin to the container"`
	T_Tty       bool   `desc:"Stdin is a TTY"`
	C_Container string `desc:"Specific container in pod, default is first container"`
}

func (e *Exec) Run(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	c, err := ctx.SpaceClient()
	if err != nil {
		return err
	}

	args := app.Args()
	if len(args) < 2 {
		return fmt.Errorf("at least two arguments are required CONTAINER CMD")
	}

	cd, err := ps.ListFirstPod(c, true, e.C_Container, args[0])
	if err != nil {
		return err
	}

	if cd == nil {
		return fmt.Errorf("failed to find pod for %s, container \"%s\"", args[0], e.C_Container)
	}

	podNS, podName, containerName := cd.Pod.Namespace, cd.Pod.Name, cd.Container.Name

	execArgs := []string{"kubectl"}
	if logrus.GetLevel() >= logrus.DebugLevel {
		execArgs = append(execArgs, "-v=9")
	}
	execArgs = append(execArgs, "-n", podNS, "exec")
	if e.I_Stdin {
		execArgs = append(execArgs, "-i")
	}
	if e.T_Tty {
		execArgs = append(execArgs, "-t")
	}

	execArgs = append(execArgs, podName, "-c", containerName)
	execArgs = append(execArgs, args[1:]...)

	logrus.Debugf("%v, KUBECONFIG=%s", execArgs, os.Getenv("KUBECONFIG"))
	cmd := reexec.Command(execArgs...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

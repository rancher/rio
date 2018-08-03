package attach

import (
	"fmt"
	"os"

	"time"

	"github.com/docker/docker/pkg/reexec"
	"github.com/rancher/rio/cli/cmd/ps"
	"github.com/rancher/rio/cli/server"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type Attach struct {
	I_Stdin bool   `desc:"Pass stdin to the container"`
	T_Tty   bool   `desc:"Stdin is a TTY"`
	Timeout string `desc:"Timeout waiting for the container to be created to attach to" default:"1m"`
}

func (a *Attach) Run(app *cli.Context) error {
	args := app.Args()
	if len(args) < 1 {
		return fmt.Errorf("at least one argument is required: CONTAINER")
	}

	timeout, err := time.ParseDuration(a.Timeout)
	if err != nil {
		return err
	}

	return RunAttach(app, timeout, a.I_Stdin, a.T_Tty, app.Args()[0])
}

func RunAttach(app *cli.Context, timeout time.Duration, stdin, tty bool, container string) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	c, err := ctx.SpaceClient()
	if err != nil {
		return err
	}

	var cd *ps.ContainerData

	deadline := time.Now().Add(timeout)
	for {
		cd, err = ps.ListFirstPod(c, true, "", container)
		if err != nil {
			return err
		}

		if (cd == nil || cd.Container == nil || cd.Container.State != "running") && time.Now().Before(deadline) {
			time.Sleep(750 * time.Millisecond)
			continue
		}

		break
	}

	if cd == nil {
		return fmt.Errorf("failed to find a container for %s", container)
	}

	podNS, podName, containerName := cd.Pod.Namespace, cd.Pod.Name, cd.Container.Name

	execArgs := []string{"kubectl"}
	if logrus.GetLevel() >= logrus.DebugLevel {
		execArgs = append(execArgs, "-v=9")
	}
	execArgs = append(execArgs, "-n", podNS, "attach")
	execArgs = append(execArgs, fmt.Sprintf("--pod-running-timeout=%s", timeout))
	execArgs = append(execArgs, podName, "-c", containerName)
	if stdin {
		execArgs = append(execArgs, "-i")
	}
	if tty {
		execArgs = append(execArgs, "-t")
	}

	logrus.Debugf("%v, KUBECONFIG=%s", execArgs, os.Getenv("KUBECONFIG"))
	cmd := reexec.Command(execArgs...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	return cmd.Run()
}

package logs

import (
	"fmt"
	"io"

	"time"

	"os"

	"github.com/docker/docker/pkg/reexec"
	"github.com/rancher/rio/cli/cmd/ps"
	"github.com/rancher/rio/cli/server"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"
)

type Logs struct {
	F_Follow    bool   `desc:"Follow log output"`
	S_Since     string `desc:"Logs since a certain time, either duration (5s, 2m, 3h) or RFC3339"`
	P_Previous  bool   `desc:"Print the logs for the previous instance of the container in a pod if it exists"`
	C_Container string `desc:"Print the logs of a specific container"`
	N_Tail      int    `desc:"Number of recent lines of logs to print, -1 for all" default:"200"`
	A_All       bool   `desc:"Include hidden or systems logs when logging"`
}

func (l *Logs) Run(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	if len(app.Args()) == 0 {
		return fmt.Errorf("at least one argument is required: CONTAINER_OR_SERVICE")
	}

	c, err := ctx.SpaceClient()
	if err != nil {
		return err
	}

	cds, err := ps.ListPods(c, l.A_All, l.C_Container, app.Args()...)
	if err != nil {
		return err
	}

	if len(cds) == 0 {
		return fmt.Errorf("failed to find container for %v, container \"%s\"", app.Args(), l.C_Container)
	}

	errg := errgroup.Group{}
	// TODO: make this suck much less.  Like color output, labels by container name, call the k8s API, not run a binary (too much overhead)
	for _, cd := range cds {
		cmd := reexec.Command("kubectl", "-n", cd.Pod.Namespace, "logs")
		if l.F_Follow {
			cmd.Args = append(cmd.Args, "-f")
		}
		if l.P_Previous {
			cmd.Args = append(cmd.Args, "-p")
		}
		if l.S_Since != "" {
			_, err := time.Parse(time.RFC3339, l.S_Since)
			if err == nil {
				cmd.Args = append(cmd.Args, "--since-time="+l.S_Since)
			} else {
				cmd.Args = append(cmd.Args, "--since="+l.S_Since)
			}
		}
		if l.N_Tail > -1 {
			cmd.Args = append(cmd.Args, fmt.Sprintf("--tail=%d", l.N_Tail))
		}

		cmd.Args = append(cmd.Args, "-c", cd.Container.Name, cd.Pod.Name)

		prefix := ""
		if len(cds) > 1 {
			prefix = fmt.Sprintf("%s/%s| ", cd.Pod.Name, cd.Container.Name)
		}
		cmd.Stdout = NewPrefixWriter(prefix, os.Stdout)
		cmd.Stderr = NewPrefixWriter(prefix, os.Stderr)
		errg.Go(func() error {
			logrus.Debugf("Running %v, KUBECONFIG=%s", cmd.Args, os.Getenv("KUBECONFIG"))
			return cmd.Run()
		})
	}

	return errg.Wait()
}

func NewPrefixWriter(prefix string, next io.Writer) io.Writer {
	return &prefixWriter{
		prefix:  []byte(prefix),
		next:    next,
		nlFound: true,
	}
}

type prefixWriter struct {
	next    io.Writer
	prefix  []byte
	nlFound bool
}

func (p *prefixWriter) Write(bytes []byte) (n int, err error) {
	np := make([]byte, 0, len(bytes))

	for _, b := range bytes {
		if p.nlFound {
			if len(np) > 0 {
				_, err := p.next.Write(np)
				if err != nil {
					return 0, err
				}
				np = make([]byte, 0, len(bytes))
			}

			_, err = p.next.Write(p.prefix)
			if err != nil {
				return 0, err
			}
		}
		p.nlFound = b == '\n'
		np = append(np, b)
	}

	if len(np) > 0 {
		_, err := p.next.Write(np)
		return len(bytes), err
	}

	return len(bytes), nil

}

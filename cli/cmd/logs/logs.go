package logs

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rancher/rio/cli/cmd/ps"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/sirupsen/logrus"
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

func (l *Logs) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return fmt.Errorf("at least one argument is required: CONTAINER_OR_SERVICE")
	}

	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}

	pds, err := ps.ListPods(ctx, l.A_All, ctx.CLI.Args()...)
	if err != nil {
		return err
	}

	if len(pds) == 0 {
		return fmt.Errorf("failed to find container for %v, container \"%s\"", ctx.CLI.Args(), l.C_Container)
	}

	errg := errgroup.Group{}
	// TODO: make this suck much less.  Like color output, labels by container name, call the k8s API, not run a binary (too much overhead)
	for _, pd := range pds {
		for _, container := range pd.Containers {
			if l.C_Container != "" && l.C_Container != container.Name {
				continue
			}

			var args []string
			if l.F_Follow {
				args = append(args, "-f")
			}
			if l.P_Previous {
				args = append(args, "-p")
			}
			if l.S_Since != "" {
				_, err := time.Parse(time.RFC3339, l.S_Since)
				if err == nil {
					args = append(args, "--since-time="+l.S_Since)
				} else {
					args = append(args, "--since="+l.S_Since)
				}
			}
			if l.N_Tail > -1 {
				args = append(args, fmt.Sprintf("--tail=%d", l.N_Tail))
			}

			args = append(args, "-c", container.Name, pd.Pod.Name)

			cmd, err := cluster.KubectlCmd(pd.Pod.Namespace, "logs", args...)
			if err != nil {
				return err
			}

			prefix := ""
			if len(pds) > 1 || len(pd.Containers) > 1 {
				prefix = fmt.Sprintf("%s/%s| ", pd.Pod.Name, container.Name)
			}
			cmd.Stdout = NewPrefixWriter(prefix, os.Stdout)
			cmd.Stderr = NewPrefixWriter(prefix, os.Stderr)
			errg.Go(func() error {
				logrus.Debugf("Running %v, KUBECONFIG=%s", cmd.Args, os.Getenv("KUBECONFIG"))
				return cmd.Run()
			})
		}
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

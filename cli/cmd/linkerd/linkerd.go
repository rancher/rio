package linkerd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/localbuilder"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	linkerdNamespace = "linkerd"

	deploymentName = "linkerd-web"

	webPort = "8084"
)

type Linkerd struct {
	Port string `desc:"The local port on which to serve requests" default:"9999"`
}

func (l *Linkerd) Customize(cmd *cli.Command) {
	cmd.Hidden = true
}

func (l *Linkerd) Run(ctx *clicontext.CLIContext) error {
	pods, err := ctx.Core.Pods(linkerdNamespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, deploymentName) {
			url := fmt.Sprintf("http://127.0.0.1:%s", l.Port)
			switch runtime.GOOS {
			case "linux":
				err = exec.Command("xdg-open", url).Start()
			case "windows":
				err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
			case "darwin":
				err = exec.Command("open", url).Start()
			default:
				err = fmt.Errorf("unsupported platform")
			}
			if err != nil {
				return err
			}
			if err := localbuilder.PortForward(ctx.K8s, l.Port, webPort, pod, true, make(chan struct{}), localbuilder.ChanWrapper(ctx.Ctx.Done())); err != nil {
				logrus.Fatal(err)
			}
			return err
		}
	}

	return fmt.Errorf("failed to find linkerd-web pods")
}

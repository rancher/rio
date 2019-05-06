package install

import (
	"fmt"
	"time"

	"github.com/rancher/rio/modules/service/controllers/serviceset"
	"github.com/sirupsen/logrus"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/pkg/systemstack"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Install struct {
	Debug     bool   `desc:"enable debug mode"`
	Namespace string `desc:"namespace to install system resources" default:"rio-system"`
}

func (i *Install) Run(ctx *clicontext.CLIContext) error {
	controllerStack := systemstack.NewStack(ctx.Apply, i.Namespace, "rio-controller", true)

	if err := controllerStack.Deploy(map[string]string{
		"NAMESPACE": i.Namespace,
		"DEBUG":     fmt.Sprint(i.Debug),
	}); err != nil {
		return err
	}
	fmt.Println("Deploying Rio control plane....")
	for {
		time.Sleep(time.Second * 2)
		dep, err := ctx.K8s.AppsV1().Deployments(ctx.SystemNamespace).Get("rio-controller", metav1.GetOptions{})
		if err != nil {
			return err
		}
		if !serviceset.IsReady(&dep.Status) {
			logrus.Debug("rio Controller is not ready yet...")
			continue
		}
		fmt.Printf("Rio control plane is deployed. Run `kubectl -n %s describe deploy rio-controller` to get more detail.\n", ctx.SystemNamespace)
		fmt.Println("Welcome to Rio!")
		break
	}
	return nil
}

package login

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/clientcfg"
	"github.com/rancher/rio/pkg/name"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
)

type Login struct {
	Kubeconfig string `desc:"KubeConfig file for k8s server"`
	Controller bool   `desc:"Running controllers"`
	NoDns      bool   `desc:"Don't run dns server"`
}

func (l *Login) Run(ctx *clicontext.CLIContext) (ex error) {
	defer func() {
		if ex == nil {
			logrus.Infof("Log in successful")
		}
	}()

	cluster := &clientcfg.Cluster{}

	restConfig, err := clientcmd.BuildConfigFromFlags("", l.Kubeconfig)
	if err != nil {
		return err
	}

	cluster.ID = name.Hex(restConfig.Host, 5)
	if err := ctx.Config.SaveCluster(cluster, restConfig); err != nil {
		return err
	}

	if l.Controller {
		context, err := runController(ctx.Ctx, l.Kubeconfig, l.NoDns)
		if err != nil {
			return err
		}
		<-context.Done()
	}
	return nil
}

package install

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/rancher/wrangler/pkg/apply"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Install struct {
	KubeConfig     string `desc:"the path to kubeconfig file" env:"KUBECONFIG"`
	Debug          bool   `desc:"enable debug mode"`
	CustomRegistry string `desc:"install controller in debug mode"`
	Namespace      string `desc:"namespace to install system resources"`
}

func (i *Install) Run(ctx *clicontext.CLIContext) error {
	restConfig, err := clientcmd.BuildConfigFromFlags("", i.KubeConfig)
	if err != nil {
		return err
	}
	k8s := kubernetes.NewForConfigOrDie(restConfig)
	apply := apply.New(k8s.Discovery(), apply.NewClientFactory(restConfig))
	controllerStack := systemstack.NewSystemStack(apply, i.Namespace, "rio-controller")

	return controllerStack.Deploy(map[string]string{
		"NAMESPACE":       i.Namespace,
		"CUSTOM_REGISTRY": i.CustomRegistry,
		"DEBUG":           fmt.Sprint(i.Debug),
	})
}

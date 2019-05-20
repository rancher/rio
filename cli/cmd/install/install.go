package install

import (
	"fmt"
	"time"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/modules/service/controllers/serviceset"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/systemstack"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Install struct {
	Debug     bool   `desc:"enable debug mode"`
	Namespace string `desc:"namespace to install system resources" default:"rio-system"`
	HTTPPort  string `desc:"http port service mesh gateway will listen to" default:"9080"`
	HTTPSPort string `desc:"https port service mesh gateway will listen to" default:"9443"`
	HostPort  bool   `desc:"whether to use hostPort to export servicemesh gateway"`
}

func (i *Install) Run(ctx *clicontext.CLIContext) error {
	if ctx.K8s == nil {
		return fmt.Errorf("can't contact Kubernetes cluster. Please make sure your cluster is accessable")
	}
	controllerStack := systemstack.NewStack(ctx.Apply, i.Namespace, "rio-controller", true)
	if _, err := ctx.Core.Namespaces().Get(i.Namespace, metav1.GetOptions{}); err != nil {
		if errors.IsNotFound(err) {
			ns := constructors.NewNamespace(i.Namespace, v1.Namespace{})
			fmt.Printf("Creating namespace %s\n", i.Namespace)
			if _, err := ctx.Core.Namespaces().Create(ns); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if err := controllerStack.Deploy(map[string]string{
		"NAMESPACE":    i.Namespace,
		"DEBUG":        fmt.Sprint(i.Debug),
		"IMAGE":        fmt.Sprintf("%s:%s", constants.ControllerImage, constants.ControllerImageTag),
		"HTTPS_PORT":   i.HTTPSPort,
		"HTTP_PORT":    i.HTTPPort,
		"USE_HOSTPORT": fmt.Sprint(i.HostPort),
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
		fmt.Printf("Rio control plane is deployed. Run `kubectl get po -n %s` to get more detail.\n", ctx.SystemNamespace)
		fmt.Println("Welcome to Rio!")
		break
	}
	return nil
}

package install

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/modules/service/controllers/serviceset"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/systemstack"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Install struct {
	HTTPPort  string   `desc:"http port service mesh gateway will listen to" default:"9080"`
	HTTPSPort string   `desc:"https port service mesh gateway will listen to" default:"9443"`
	HostPorts bool     `desc:"whether to use hostPorts to expose service mesh gateway"`
	IPAddress []string `desc:"Manually specify IP addresses to generate rdns domain"`
}

func (i *Install) Run(ctx *clicontext.CLIContext) error {
	if ctx.K8s == nil {
		return fmt.Errorf("can't contact Kubernetes cluster. Please make sure your cluster is accessable")
	}

	namespace := ctx.SystemNamespace
	if namespace == "" {
		namespace = "rio-system"
	}

	controllerStack := systemstack.NewStack(ctx.Apply, namespace, "rio-controller", true)
	if _, err := ctx.Core.Namespaces().Get(namespace, metav1.GetOptions{}); err != nil {
		if errors.IsNotFound(err) {
			ns := constructors.NewNamespace(namespace, v1.Namespace{})
			fmt.Printf("Creating namespace %s\n", namespace)
			if _, err := ctx.Core.Namespaces().Create(ns); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// hack for detecting minikube cluster
	nodes, err := ctx.Core.Nodes().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(nodes.Items) == 1 && nodes.Items[0].Name == "minikube" {
		fmt.Println("Detecting that you are using minikube cluster")
		cmd := exec.Command("minikube", "ip")
		stdout := &strings.Builder{}
		stderr := &strings.Builder{}
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("$(minikube ip) failed with error: (%v). Do you have minikube in your PATH", stderr.String())
		}
		ip := strings.Trim(stdout.String(), " ")
		fmt.Printf("Manually setting minikube IP to %s\n", ip)
		i.IPAddress = []string{ip}
		i.HostPorts = true
	}

	if err := controllerStack.Deploy(map[string]string{
		"NAMESPACE":    namespace,
		"DEBUG":        fmt.Sprint(ctx.Debug),
		"IMAGE":        fmt.Sprintf("%s:%s", constants.ControllerImage, constants.ControllerImageTag),
		"HTTPS_PORT":   i.HTTPSPort,
		"HTTP_PORT":    i.HTTPPort,
		"USE_HOSTPORT": fmt.Sprint(i.HostPorts),
		"IP_ADDRESSES": strings.Join(i.IPAddress, ","),
	}); err != nil {
		return err
	}
	fmt.Println("Deploying Rio control plane....")
	for {
		time.Sleep(time.Second * 2)
		dep, err := ctx.K8s.AppsV1().Deployments(namespace).Get("rio-controller", metav1.GetOptions{})
		if err != nil {
			return err
		}
		if !serviceset.IsReady(&dep.Status) {
			fmt.Printf("Waiting for deployment %s/%s to become ready\n", dep.Namespace, dep.Name)
			continue
		}
		info, err := ctx.Project.RioInfos().Get("rio", metav1.GetOptions{})
		if err != nil {
			fmt.Println("Waiting for rio controller to initialize")
			continue
		} else if info.Status.Version == "" {
			fmt.Println("Waiting for rio controller to initialize")
			continue
		} else {
			fmt.Printf("rio controller version %s (%s) installed into namespace %s\n", info.Status.Version, info.Status.GitCommit, info.Status.SystemNamespace)
		}
		fmt.Printf("Rio control plane is deployed. Run `kubectl get po -n %s` to get more detail.\n", info.Status.SystemNamespace)
		fmt.Println("Controller logs are available from `rio systemlogs`")
		fmt.Println("")
		fmt.Println("Welcome to Rio!")
		break
	}
	return nil
}

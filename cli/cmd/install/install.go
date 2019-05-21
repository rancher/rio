package install

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/modules/service/controllers/serviceset"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/systemstack"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	SystemComponents = []string{
		Autoscaler,
		BuildController,
		Buildkit,
		CertManager,
		Grafana,
		IstioCitadel,
		IstioPilot,
		IstioTelemetry,
		Kiali,
		Prometheus,
		Registry,
		Webhook,
	}

	Autoscaler      = "autoscaler"
	BuildController = "build-controller"
	Buildkit        = "buildkit"
	CertManager     = "cert-manager"
	Grafana         = "grafana"
	IstioCitadel    = "istio-citadel"
	IstioPilot      = "istio-pilot"
	IstioTelemetry  = "istio-telemetry"
	Kiali           = "kiali"
	Prometheus      = "prometheus"
	Registry        = "registry"
	Webhook         = "webhook"
)

type Install struct {
	HTTPPort    string   `desc:"http port service mesh gateway will listen to" default:"9080"`
	HTTPSPort   string   `desc:"https port service mesh gateway will listen to" default:"9443"`
	HostPorts   bool     `desc:"whether to use hostPorts to expose service mesh gateway"`
	IPAddress   []string `desc:"Manually specify IP addresses to generate rdns domain"`
	ServiceCidr string   `desc:"Manually specify service CIDR for service mesh to intercept"`
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

	memoryWarning := false
	var totalMemory int64
	for _, node := range nodes.Items {
		totalMemory += node.Status.Capacity.Memory().Value()
	}
	if totalMemory < 2147000000 {
		memoryWarning = true
	}

	if isMinikubeCluster(nodes) && len(i.IPAddress) == 0 {
		fmt.Println("Detected minikube cluster")
		cmd := exec.Command("minikube", "ip")
		stdout := &strings.Builder{}
		stderr := &strings.Builder{}
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("$(minikube ip) failed with error: (%v). Do you have minikube in your PATH", stderr.String())
		}
		ip := strings.Trim(stdout.String(), " ")
		fmt.Printf("Manually setting cluster IP to %s\n", ip)
		i.IPAddress = []string{ip}
		i.HostPorts = true
	}

	if memoryWarning {
		if isMinikubeCluster(nodes) {
			fmt.Println("Warning: detecting that your minikube cluster doesn't have at least 3 GB of memory. Please try to increase memory by running `minikube start --memory 4098`")
		} else if isDockerForMac(nodes) {
			fmt.Println("Warning: detecting that your Docker For Mac cluster doesn't have at least 3 GB of memory. Please try to increase memory by following the doc https://docs.docker.com/v17.12/docker-for-mac.")
		} else {
			fmt.Println("Warning: detecting that your cluster doesn't have at least 3 GB of memory in total. Please try to increase memory for your nodes")
		}
	}

	if i.ServiceCidr == "" {
		svc, err := ctx.Core.Services("default").Get("kubernetes", metav1.GetOptions{})
		if err != nil {
			return err
		}
		clusterCIDR := svc.Spec.ClusterIP + "/16"
		fmt.Printf("Defaulting cluster CIDR to %s\n", clusterCIDR)
		i.ServiceCidr = clusterCIDR
	}

	if err := controllerStack.Deploy(map[string]string{
		"NAMESPACE":    namespace,
		"DEBUG":        fmt.Sprint(ctx.Debug),
		"IMAGE":        fmt.Sprintf("%s:%s", constants.ControllerImage, constants.ControllerImageTag),
		"HTTPS_PORT":   i.HTTPSPort,
		"HTTP_PORT":    i.HTTPPort,
		"USE_HOSTPORT": fmt.Sprint(i.HostPorts),
		"IP_ADDRESSES": strings.Join(i.IPAddress, ","),
		"SERVICE_CIDR": i.ServiceCidr,
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
		} else if notReadyList, ok := allReady(info); !ok {
			fmt.Printf("Waiting for all the system components to be up. Not ready component: %v\n", notReadyList)
			time.Sleep(15 * time.Second)
			continue
		} else {
			fmt.Printf("rio controller version %s (%s) installed into namespace %s\n", info.Status.Version, info.Status.GitCommit, info.Status.SystemNamespace)
		}
		fmt.Printf("Please make sure all the system pods are actually running. Run `kubectl get po -n %s` to get more detail.\n", info.Status.SystemNamespace)
		fmt.Println("Controller logs are available from `rio systemlogs`")
		fmt.Println("")
		fmt.Println("Welcome to Rio!")
		fmt.Println("")
		fmt.Println("Run `rio run https://github.com/rancher/rio-demo` as an example")
		break
	}
	return nil
}

func isMinikubeCluster(nodes *v1.NodeList) bool {
	return len(nodes.Items) == 1 && nodes.Items[0].Name == "minikube"
}

func isDockerForMac(nodes *v1.NodeList) bool {
	return len(nodes.Items) == 1 && nodes.Items[0].Name == "docker-for-desktop"
}

func allReady(info *adminv1.RioInfo) ([]string, bool) {
	var notReadyList []string
	ready := true
	for _, c := range SystemComponents {
		if info.Status.SystemComponentReadyMap[c] != "running" {
			notReadyList = append(notReadyList, c)
			ready = false
		}
	}
	return notReadyList, ready
}

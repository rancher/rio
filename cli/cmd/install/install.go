package install

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rancher/mapper/slice"

	"github.com/docker/docker/pkg/reexec"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/up/questions"
	"github.com/rancher/rio/modules/service/controllers/serviceset"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/systemstack"
	v1 "k8s.io/api/core/v1"
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

	featureMap = map[string]string{
		Autoscaler:      "autoscaling",
		BuildController: "build",
		Buildkit:        "build",
		CertManager:     "letsencrypt",
		Grafana:         "grafana",
		IstioCitadel:    "istio",
		IstioPilot:      "istio",
		IstioTelemetry:  "mixer",
		Kiali:           "kiali",
		Prometheus:      "prometheus",
		Webhook:         "build",
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

	statusIndex = 0
	statusChar  = "|/-\\"
)

type Install struct {
	HTTPPort        string   `desc:"Http port service mesh gateway will listen to" default:"9080" name:"http-port"`
	HTTPSPort       string   `desc:"Https port service mesh gateway will listen to" default:"9443" name:"https-port"`
	HostPorts       bool     `desc:"Whether to use hostPorts to expose service mesh gateway"`
	IPAddress       []string `desc:"Manually specify IP addresses to generate rdns domain" name:"ip-address"`
	ServiceCidr     string   `desc:"Manually specify service CIDR for service mesh to intercept"`
	DisableFeatures []string `desc:"Manually specify features to disable, supports CSV"`
	Yaml            bool     `desc:"Only print out k8s yaml manifest"`
	Check           bool     `desc:"Only check status, don't deploy controller'"`
	Lite            bool     `desc:"Only install lite version of Rio(monitoring will be disabled, will be ignored if --disable-features is set)"`
}

func (i *Install) Run(ctx *clicontext.CLIContext) error {
	if ctx.K8s == nil {
		return fmt.Errorf("Can't contact Kubernetes cluster. Please make sure your cluster is accessible")
	}
	out := os.Stdout
	if i.Yaml {
		devnull, err := os.Open(os.DevNull)
		if err != nil {
			return err
		}
		out = devnull
	}

	namespace := ctx.SystemNamespace
	if namespace == "" {
		namespace = "rio-system"
		ctx.SystemNamespace = namespace
	}

	controllerStack := systemstack.NewStack(ctx.Apply, namespace, "rio-controller", true)

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
		fmt.Fprintln(out, "Detected minikube cluster")
		cmd := exec.Command("minikube", "ip")
		stdout := &strings.Builder{}
		stderr := &strings.Builder{}
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("$(minikube ip) failed with error: (%v). Do you have minikube in your PATH", stderr.String())
		}
		ip := strings.Trim(stdout.String(), " ")
		fmt.Fprintf(out, "Manually setting cluster IP to %s\n", ip)
		i.IPAddress = []string{ip}
		i.HostPorts = true
	}

	if memoryWarning {
		if isMinikubeCluster(nodes) {
			fmt.Fprintln(out, "Warning: detecting that your minikube cluster doesn't have at least 3 GB of memory. Please try to increase memory by running `minikube start --memory 4098`")
		} else if isDockerForMac(nodes) {
			fmt.Fprintln(out, "Warning: detecting that your Docker For Mac cluster doesn't have at least 3 GB of memory. Please try to increase memory by following the doc https://docs.docker.com/v17.12/docker-for-mac.")
		} else {
			fmt.Fprintln(out, "Warning: detecting that your cluster doesn't have at least 3 GB of memory in total. Please try to increase memory for your nodes")
		}
		i.Lite = true
	}

	if i.ServiceCidr == "" {
		svc, err := ctx.Core.Services("default").Get("kubernetes", metav1.GetOptions{})
		if err != nil {
			return err
		}
		clusterCIDR := svc.Spec.ClusterIP + "/16"
		fmt.Fprintf(out, "Defaulting cluster CIDR to %s\n", clusterCIDR)
		i.ServiceCidr = clusterCIDR
	}

	if i.Lite && len(i.DisableFeatures) == 0 {
		fmt.Fprintf(out, "Setting install mode to lite, monitoring features will be disabled\n")
		i.DisableFeatures = []string{"mixer", "grafana", "kiali", "prometheus"}
	}
	answers := map[string]string{
		"NAMESPACE":        namespace,
		"DEBUG":            fmt.Sprint(ctx.Debug),
		"IMAGE":            fmt.Sprintf("%s:%s", constants.ControllerImage, constants.ControllerImageTag),
		"HTTPS_PORT":       i.HTTPSPort,
		"HTTP_PORT":        i.HTTPPort,
		"USE_HOSTPORT":     fmt.Sprint(i.HostPorts),
		"IP_ADDRESSES":     strings.Join(i.IPAddress, ","),
		"SERVICE_CIDR":     i.ServiceCidr,
		"DISABLE_FEATURES": strings.Join(i.DisableFeatures, ","),
	}
	if i.Yaml {
		yamlOutput, err := controllerStack.Yaml(answers)
		if err != nil {
			return err
		}
		fmt.Println(yamlOutput)
		return nil
	}

	if !i.Check {
		fmt.Println("Deploying Rio control plane....")
		if err := controllerStack.Deploy(answers); err != nil {
			return err
		}
	}

	start := time.Now()
	for {
		time.Sleep(time.Second * 2)
		if !i.Check {
			dep, err := ctx.K8s.AppsV1().Deployments(namespace).Get("rio-controller", metav1.GetOptions{})
			if err != nil {
				return err
			}
			if !serviceset.IsReady(&dep.Status) {
				fmt.Printf("\r%v Waiting for deployment %s/%s to become ready", string(statusChar[statusIndex]), dep.Namespace, dep.Name)
				statusIndex = modIndex(statusIndex)
				continue
			}
		}
		info, err := ctx.Project.RioInfos().Get("rio", metav1.GetOptions{})
		if err != nil {
			fmt.Printf("\r%v Waiting for rio controller to initialize", string(statusChar[statusIndex]))
			statusIndex = modIndex(statusIndex)
			time.Sleep(time.Second)
			continue
		} else if info.Status.Version == "" {
			fmt.Printf("\r%v Waiting for rio controller to initialize", string(statusChar[statusIndex]))
			statusIndex = modIndex(statusIndex)
			continue
		} else if notReadyList, ok := allReady(info, i.DisableFeatures); !ok {
			fmt.Printf("\r%v Waiting for all the system components to be up. Not ready: %v", string(statusChar[statusIndex]), notReadyList)
			statusIndex = modIndex(statusIndex)
			time.Sleep(2 * time.Second)
			continue
		} else {
			ok, err := i.delectingServiceLoadbalancer(ctx, info, start)
			if err != nil {
				return err
			} else if !ok {
				fmt.Printf("\r%v Waiting for service loadbalancer to be up", string(statusChar[statusIndex]))
				statusIndex = modIndex(statusIndex)
				time.Sleep(2 * time.Second)
				continue
			}
			fmt.Printf("\rrio controller version %s (%s) installed into namespace %s\n", info.Status.Version, info.Status.GitCommit, info.Status.SystemNamespace)
		}

		fmt.Println("Detecting if clusterDomain is accessible...")
		clusterDomain, err := ctx.Domain()
		if err != nil {
			return err
		}
		if clusterDomain == "" {
			fmt.Printf("Warning: Detected that Rio cluster domain is not generated for this cluster right now")
		} else {
			_, err = http.Get(fmt.Sprintf("http://%s:%s", clusterDomain, i.HTTPPort))
			if err != nil {
				fmt.Printf("Warning: ClusterDomain is not accessible. Error: %v\n", err)
			} else {
				fmt.Println("ClusterDomain is reachable. Run `rio info` to get more info.")
			}
		}

		fmt.Println("Controller logs are available from `rio systemlogs`")
		fmt.Println("")
		fmt.Println("Welcome to Rio!")
		fmt.Println("")
		fmt.Println("Run `rio run https://github.com/rancher/rio-demo` as an example")
		break
	}
	return nil
}

func modIndex(v int) int {
	v++
	return int(math.Mod(float64(v), 4.0))
}

func isMinikubeCluster(nodes *v1.NodeList) bool {
	return len(nodes.Items) == 1 && nodes.Items[0].Name == "minikube"
}

func isDockerForMac(nodes *v1.NodeList) bool {
	return len(nodes.Items) == 1 && nodes.Items[0].Name == "docker-for-desktop"
}

func allReady(info *adminv1.RioInfo, disabledFeatures []string) ([]string, bool) {
	var notReadyList []string
	ready := true
	for _, c := range SystemComponents {
		if !slice.ContainsString(disabledFeatures, featureMap[c]) {
			if info.Status.SystemComponentReadyMap[c] != "running" {
				notReadyList = append(notReadyList, c)
				ready = false
			}
		}
	}
	return notReadyList, ready
}

func (i *Install) delectingServiceLoadbalancer(ctx *clicontext.CLIContext, info *adminv1.RioInfo, startTime time.Time) (bool, error) {
	svc, err := ctx.Core.Services(info.Status.SystemNamespace).Get(fmt.Sprintf("%s-%s", constants.IstioGateway, constants.DefaultServiceVersion), metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	if svc.Spec.Type == v1.ServiceTypeLoadBalancer && !i.HostPorts {
		if len(svc.Status.LoadBalancer.Ingress) == 0 || (svc.Status.LoadBalancer.Ingress[0].Hostname != "" && svc.Status.LoadBalancer.Ingress[0].Hostname != "localhost") {
			if time.Now().After(startTime.Add(time.Minute * 2)) {
				msg := ""
				if len(svc.Status.LoadBalancer.Ingress) > 0 {
					msg = fmt.Sprintln("\nDetecting that your service loadbalancer generates a DNS endpoint(usually AWS provider). Rio doesn't support it right now. Do you want to:")
				} else {
					msg = fmt.Sprintln("\nDetecting that your service loadbalancer for service mesh gateway is still pending. Do you want to:")
				}

				options := []string{
					fmt.Sprintf("[1] Use HostPorts(Please make sure port %v and %v are open for your nodes)\n", i.HTTPPort, i.HTTPSPort),
					"[2] Wait for Service Load Balancer\n",
				}

				num, err := questions.PromptOptions(msg, -1, options...)
				if err != nil {
					return false, nil
				}

				if num == 0 {
					fmt.Println("Reinstall Rio using --host-ports")
					cmd := reexec.Command("rio", "install", "--host-ports", "--http-port", i.HTTPPort, "--https-port", i.HTTPSPort)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					cmd.Env = os.Environ()
					return true, cmd.Run()
				}
				return true, nil
			}
			return false, nil
		}
	}
	return true, nil
}

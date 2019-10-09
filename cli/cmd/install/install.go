package install

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/docker/docker/pkg/reexec"
	"github.com/rancher/mapper/slice"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/progress"
	"github.com/rancher/rio/cli/pkg/up/questions"
	"github.com/rancher/rio/modules/service/controllers/serviceset"
	adminv1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/pkg/version"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	SystemComponents = []string{
		Autoscaler,
		BuildController,
		CertManager,
		Gateway,
		IstioPilot,
		Registry,
		Webhook,
	}
	istioComponents = []string{
		IstioGrafana,
		IstioCitadel,
		IstioTelemetry,
		IstioKiali,
		IstioPrometheus,
		IstioSidecarInjector,
	}

	featureMap = map[string]string{
		Autoscaler: constants.FeatureAutoscaling,

		CertManager: constants.FeatureLetsencrypts,

		Gateway:    constants.FeatureGateway,
		IstioPilot: constants.FeatureGateway,

		IstioSidecarInjector: constants.FeatureIstio,
		IstioCitadel:         constants.FeatureIstio,
		IstioGrafana:         constants.FeatureGrafana,
		IstioTelemetry:       constants.FeatureMixer,
		IstioKiali:           constants.FeatureKiali,
		IstioPrometheus:      constants.FeaturePromethues,

		BuildController: constants.FeatureBuild,
		Registry:        constants.FeatureBuild,
		Webhook:         constants.FeatureBuild,
	}

	Autoscaler           = "autoscaler"
	BuildController      = "build-controller"
	CertManager          = "cert-manager"
	IstioGrafana         = "grafana"
	IstioCitadel         = "istio-citadel"
	IstioPilot           = "istio-pilot"
	IstioTelemetry       = "istio-telemetry"
	IstioKiali           = "kiali"
	IstioSidecarInjector = "istio-sidecar-injector"
	IstioPrometheus      = "prometheus"
	Gateway              = "gateway"
	Registry             = "registry"
	Webhook              = "webhook"
)

type Install struct {
	HTTPPort        string   `desc:"Http port service mesh gateway will listen to" default:"9080" name:"http-port"`
	HTTPSPort       string   `desc:"Https port service mesh gateway will listen to" default:"9443" name:"https-port"`
	IPAddress       []string `desc:"Manually specify IP addresses to generate rdns domain, supports comma separated values" name:"ip-address"`
	DisableFeatures []string `desc:"Manually specify features to disable, supports comma separated values"`
	HTTPProxy       string   `desc:"Set HTTP_PROXY environment variable for control plane"`
	Yaml            bool     `desc:"Only print out k8s yaml manifest"`
	Check           bool     `desc:"Only check status, don't deploy controller"`
	Lite            bool     `desc:"Only install lite version of Rio istio install(only works if mesh-mode is istio, monitoring will be disabled, will be ignored if --disable-features is set)"`
	Mode            string   `desc:"Install mode to expose gateway. Available options are svclb and hostport" default:"svclb"`
	MeshMode        string   `desc:"Service mesh mode. Options: (linkerd/istio)" default:"linkerd"`
}

func (i *Install) Run(ctx *clicontext.CLIContext) error {
	if ctx.K8s == nil {
		return fmt.Errorf("can't contact Kubernetes cluster. Please make sure your cluster is accessible")
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

	controllerStack := stack.NewSystemStack(ctx.Apply, namespace, "rio-controller")

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
		ip := strings.Trim(stdout.String(), " \n")
		fmt.Fprintf(out, "Manually setting cluster IP to %s and install mode to %s\n", ip, constants.InstallModeHostport)
		i.IPAddress = []string{ip}
		i.Mode = constants.InstallModeHostport
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

	if i.Lite && len(i.DisableFeatures) == 0 {
		fmt.Fprintf(out, "Setting install mode to lite, monitoring features will be disabled\n")
		i.DisableFeatures = []string{IstioTelemetry, IstioGrafana, IstioKiali, IstioPrometheus}
	}

	if i.Mode == constants.InstallModeIngress {
		i.HTTPPort = "80"
		i.HTTPSPort = "443"
	}

	answers := map[string]string{
		"NAMESPACE":        namespace,
		"DEBUG":            fmt.Sprint(ctx.Debug),
		"IMAGE":            fmt.Sprintf("%s:%s", constants.ControllerImage, constants.ControllerImageTag),
		"HTTPS_PORT":       i.HTTPSPort,
		"HTTP_PORT":        i.HTTPPort,
		"INSTALL_MODE":     i.Mode,
		"IP_ADDRESSES":     strings.Join(i.IPAddress, ","),
		"DISABLE_FEATURES": strings.Join(i.DisableFeatures, ","),
		"HTTP_PROXY":       i.HTTPProxy,
		"SM_MODE":          i.MeshMode,
		"RUN_CONTROLLER":   "true",
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

	var disabledFeatures []string
	for _, dfs := range i.DisableFeatures {
		parts := strings.Split(dfs, ",")
		for _, p := range parts {
			disabledFeatures = append(disabledFeatures, strings.Trim(p, " "))
		}
	}
	i.DisableFeatures = disabledFeatures
	if i.MeshMode == "linkerd" {
		disabledFeatures = append(disabledFeatures, "istio")
	} else if i.MeshMode == "istio" {
		disabledFeatures = append(disabledFeatures, "linkerd")
	}

	progress := progress.NewWriter()
	start := time.Now()
	for {
		// Checking rio-controller deployment
		if !i.Check {
			dep, err := ctx.K8s.AppsV1().Deployments(namespace).Get("rio-controller", metav1.GetOptions{})
			if err != nil {
				return err
			}
			if !serviceset.IsReady(&dep.Status) {
				progress.Display("Waiting for deployment %s/%s to become ready", 2, dep.Namespace, dep.Name)
				continue
			}
		}

		// Checking systemInfo and components
		info, err := ctx.Project.RioInfos().Get("rio", metav1.GetOptions{})
		if err != nil || info.Status.Version == "" {
			progress.Display("Waiting for rio controller to initialize", 2)
			continue
		} else if notReadyList, ok := allReady(info, i.MeshMode, disabledFeatures); !ok {
			progress.Display("Waiting for all the system components to be up. Not ready: %v", 2, notReadyList)
			continue
		} else {
			ok, err := i.fallbackInstall(ctx, info, start)
			if err != nil {
				return err
			} else if !ok {
				progress.Display("Waiting for service loadbalancer to be up", 2)
				continue
			}
			fmt.Printf("\rrio controller version %s (%s) installed into namespace %s\n", version.Version, version.GitCommit, info.Status.SystemNamespace)
		}

		// Checking if clusterDomain is available
		fmt.Println("Detecting if clusterDomain is accessible...")
		clusterDomain, err := ctx.Domain()
		if err != nil {
			return err
		}
		if clusterDomain == "" {
			fmt.Println("Warning: Detected that Rio cluster domain is not generated for this cluster right now")
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

func isMinikubeCluster(nodes *v1.NodeList) bool {
	return len(nodes.Items) == 1 && nodes.Items[0].Name == "minikube"
}

func isDockerForMac(nodes *v1.NodeList) bool {
	return len(nodes.Items) == 1 && nodes.Items[0].Name == "docker-for-desktop"
}

type list struct {
	notReady []string
}

func (l list) String() string {
	sort.Strings(l.notReady)
	if len(l.notReady) > 3 {
		return fmt.Sprint(append(l.notReady[:3], "..."))
	}
	return fmt.Sprint(l.notReady)
}

func allReady(info *adminv1.RioInfo, smMode string, disabledFeatures []string) (list, bool) {
	var l list
	ready := true
	components := SystemComponents
	if smMode == "istio" {
		components = append(components, istioComponents...)
	}
	for _, c := range components {
		if !slice.ContainsString(disabledFeatures, featureMap[c]) {
			if info.Status.SystemComponentReadyMap[c] != "running" {
				l.notReady = append(l.notReady, c)
				ready = false
			}
		}
	}
	return l, ready
}

func (i *Install) fallbackInstall(ctx *clicontext.CLIContext, info *adminv1.RioInfo, startTime time.Time) (bool, error) {
	if i.Mode == constants.InstallModeIngress {
		ingress, err := ctx.K8s.NetworkingV1beta1().Ingresses(info.Status.SystemNamespace).Get(constants.ClusterIngressName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if len(ingress.Status.LoadBalancer.Ingress) == 0 || (ingress.Status.LoadBalancer.Ingress[0].Hostname != "" && ingress.Status.LoadBalancer.Ingress[0].Hostname != "localhost") {
			if time.Now().After(startTime.Add(time.Minute * 2)) {
				msg := ""
				if len(ingress.Status.LoadBalancer.Ingress) > 0 {
					msg = fmt.Sprintln("\nDetecting that your ingress generates a DNS endpoint(usually AWS provider). Rio doesn't support it right now. Do you want to:")
				} else {
					msg = fmt.Sprintln("\nDetecting that your ingress for service mesh gateway is still pending. Do you want to:")
				}

				options := []string{
					"[1]: Use Service Loadbalancer\n",
					fmt.Sprintf("[2]: Use HostPorts (Please make sure port %v and %v are open for your nodes)\n", i.HTTPPort, i.HTTPSPort),
					"[3]: Wait for ingress\n",
				}

				num, err := questions.PromptOptions(msg, -1, options...)
				if err != nil {
					return false, nil
				}

				if num == 0 {
					fmt.Println("Reinstall Rio using svclb")
					return true, i.reinstall(constants.InstallModeSvclb)
				}

				if num == 1 {
					fmt.Println("Reinstall Rio using hostport")
					return true, i.reinstall(constants.InstallModeHostport)
				}
				return true, nil
			}
			return false, nil
		}
	}

	if i.Mode == constants.InstallModeSvclb {
		svc, err := ctx.Core.Services(info.Status.SystemNamespace).Get(fmt.Sprintf("%s-%s", constants.GatewayName, constants.DefaultServiceVersion), metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if len(svc.Status.LoadBalancer.Ingress) == 0 || (svc.Status.LoadBalancer.Ingress[0].Hostname != "" && svc.Status.LoadBalancer.Ingress[0].Hostname != "localhost") {
			if time.Now().After(startTime.Add(time.Minute * 2)) {
				msg := ""
				if len(svc.Status.LoadBalancer.Ingress) > 0 {
					msg = fmt.Sprintln("\nDetecting that your service loadbalancer generates a DNS endpoint(usually AWS provider). Rio doesn't support it right now. Do you want to:")
				} else {
					msg = fmt.Sprintln("\nDetecting that your service loadbalancer for service mesh gateway is still pending. Do you want to:")
				}

				options := []string{
					fmt.Sprintf("[1]: Use HostPorts (Please make sure port %v and %v are open for your nodes)\n", i.HTTPPort, i.HTTPSPort),
					"[2]: Wait for Service Load Balancer\n",
				}

				num, err := questions.PromptOptions(msg, -1, options...)
				if err != nil {
					return false, nil
				}

				if num == 0 {
					fmt.Println("Reinstall Rio using hostport")
					return true, i.reinstall(constants.InstallModeHostport)
				}
				return true, nil
			}
			return false, nil
		}
	}
	return true, nil
}

func (i *Install) reinstall(mode string) error {
	args := []string{"rio", "install", "--mode", mode, "--http-port", i.HTTPPort, "--https-port", i.HTTPSPort}
	for _, ip := range i.IPAddress {
		args = append(args, "--ip-address", ip)
	}
	for _, df := range i.DisableFeatures {
		args = append(args, "--disable-features", df)
	}
	if i.HTTPProxy != "" {
		args = append(args, "--httpproxy", i.HTTPProxy)
	}
	cmd := reexec.Command(args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	return cmd.Run()
}

package localbuilder

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/rancher/rio/pkg/generated/clientset/versioned/scheme"
	"github.com/rancher/wrangler/pkg/kubeconfig"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	portforwardtools "k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/kubectl/cmd/portforward"
)

func portForward(podName, podNamespace string, k8s *kubernetes.Clientset, port, targetPort string, stopChan chan struct{}) error {
	loader := kubeconfig.GetInteractiveClientConfig(os.Getenv("KUBECONFIG"))

	restConfig, err := loader.ClientConfig()
	if err != nil {
		return err
	}
	if err := setConfigDefaults(restConfig); err != nil {
		return err
	}
	restClient, err := rest.RESTClientFor(restConfig)
	if err != nil {
		return err
	}
	ioStreams := genericclioptions.IOStreams{}

	portForwardOpt := portforward.PortForwardOptions{
		Namespace:    podNamespace,
		PodName:      podName,
		RESTClient:   restClient,
		Config:       restConfig,
		PodClient:    k8s.CoreV1(),
		Address:      []string{"localhost"},
		Ports:        []string{fmt.Sprintf("%s:%s", port, targetPort)},
		StopChannel:  stopChan,
		ReadyChannel: make(chan struct{}),
		PortForwarder: &defaultPortForwarder{
			IOStreams: ioStreams,
		},
	}
	return portForwardOpt.RunPortForward()
}

type defaultPortForwarder struct {
	genericclioptions.IOStreams
}

func (f *defaultPortForwarder) ForwardPorts(method string, url *url.URL, opts portforward.PortForwardOptions) error {
	transport, upgrader, err := spdy.RoundTripperFor(opts.Config)
	if err != nil {
		return err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, method, url)
	fw, err := portforwardtools.NewOnAddresses(dialer, opts.Address, opts.Ports, opts.StopChannel, opts.ReadyChannel, f.Out, f.ErrOut)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/api"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

func isReady(status *appv1.DeploymentStatus) bool {
	if status == nil {
		return false
	}
	for _, con := range status.Conditions {
		if con.Type == appv1.DeploymentAvailable && con.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

func findPod(clientset *kubernetes.Clientset, namespace string, selector string) (*v1.Pod, error) {
	pods, err := clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, err
	}
	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pod found")
	}
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning {
			return &pod, nil
		}
	}
	return &pods.Items[0], nil
}

package kubeconfig

import (
	"io"
	"os"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
)

func GetNonInteractiveClientConfig(kubeConfig string) clientcmd.ClientConfig {
	return GetClientConfig(kubeConfig, nil)
}

func GetInteractiveClientConfig(kubeConfig string) clientcmd.ClientConfig {
	return GetClientConfig(kubeConfig, os.Stdin)
}

func GetClientConfig(kubeConfig string, reader io.Reader) clientcmd.ClientConfig {
	loadingRules := GetLoadingRules(kubeConfig)
	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}
	return clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, overrides, reader)
}

func GetLoadingRules(kubeConfig string) *clientcmd.ClientConfigLoadingRules {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	if kubeConfig != "" {
		loadingRules.ExplicitPath = kubeConfig
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		loadingRules.Precedence = append(loadingRules.Precedence, filepath.Join(homeDir, ".kube", "k3s.yaml"))
	}

	loadingRules.Precedence = append(loadingRules.Precedence, "/etc/rancher/k3s/k3s.yaml")
	return loadingRules
}

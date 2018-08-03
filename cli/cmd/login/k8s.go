package login

import (
	"errors"
	"fmt"
	"strconv"

	"time"

	"github.com/rancher/rio/cli/pkg/up/questions"
	"github.com/rancher/rio/cli/server"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func (l *Login) k8s(tempFile string) (bool, error) {
	if l.S_Server != "" || l.T_Token != "" {
		return false, nil
	}

	k8s, err := questions.PromptBool("Do you want to try to install Rio on an existing Kubernetes installation?", false)
	if err != nil {
		return false, err
	}
	if !k8s {
		return false, nil
	}

	clientConfig, err := l.tryKubernetes()
	if err != nil {
		return false, err
	}
	if clientConfig == nil {
		return false, errors.New("no valid kubernetes kubeconfig was found, try using --kubeconfig option")
	}

	err = InstallRioInK8s(clientConfig)
	if err != nil {
		return false, err
	}

	cfg, err := clientConfig.RawConfig()
	if err != nil {
		return false, err
	}

	defer func() {
		for i := 0; i < 60; i++ {
			_, err := server.SpaceClient(tempFile, k8s)
			if err == nil {
				return
			}
			if i == 1 {
				logrus.Infof("Waiting to connect to Rio")
			}
			time.Sleep(2 * time.Second)
		}
	}()

	return true, clientcmd.WriteToFile(cfg, tempFile)
}

func (l *Login) tryKubernetes() (clientcmd.ClientConfig, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.ExplicitPath = l.Kubeconfig
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules,
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}})
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, nil
	}

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, nil
	}

	info, err := client.Discovery().ServerVersion()
	if err != nil {
		return nil, nil
	}

	if !is110OrGreater(*info) {
		logrus.Infof("Found Kubernetes v%s.%s but v1.10 or newer is required", info.Major, info.Minor)
		return nil, nil
	}

	ok, err := questions.PromptBool(fmt.Sprintf("You seem to have Kubernetes v%s.%s available at %s, would you like to use that for Rio?",
		info.Major, info.Minor, restConfig.Host), false)
	if err != nil || !ok {
		return nil, err
	}

	client.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rio-system",
		},
	})

	return clientConfig, nil
}

func is110OrGreater(info version.Info) bool {
	if info.Major != "1" {
		return false
	}

	minor, err := strconv.Atoi(info.Minor)
	if err != nil {
		return false
	}
	return minor >= 10
}

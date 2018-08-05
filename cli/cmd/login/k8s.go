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
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func (l *Login) k8s(tempFile string) error {
	clientConfig, err := l.testKubernetes(tempFile)
	if err != nil {
		return err
	}
	if clientConfig == nil {
		return errors.New("no valid kubernetes kubeconfig was found, try using --kubeconfig option")
	}

	defer func() {
		for i := 0; i < 60; i++ {
			_, err := server.SpaceClient(tempFile, true)
			if err == nil {
				return
			}
			if i == 1 {
				logrus.Infof("Waiting to connect to Rio")
			}
			time.Sleep(2 * time.Second)
		}
	}()

	return InstallRioInK8s(clientConfig)
}

func loadKubeConfig(defConfig string) (clientcmd.ClientConfig, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.ExplicitPath = defConfig
	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules,
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}})
	clientConfig, err := cc.RawConfig()
	if err == nil && len(clientConfig.Contexts) == 0 {
		err = errors.New("failed to find a valid Kubernetes configuration")
	}
	return cc, err
}

func (l *Login) selectContextAndWriteToFile(tempFile string) error {
	cc, err := loadKubeConfig(l.Kubeconfig)
	if err != nil {
		return err
	}

	rawConfig, err := cc.RawConfig()
	if len(rawConfig.Contexts) > 1 {
		var options []string
		optionIndex := map[int]string{}

		i := 0
		for name := range rawConfig.Contexts {
			optionIndex[i] = name
			i++
			options = append(options, fmt.Sprintf("[%d] %s\n", i, name))
		}

		num, err := questions.PromptOptions("Select which context to use\n", 1, options...)
		if err != nil {
			return err
		}

		rawConfig.CurrentContext = optionIndex[num]
	}

	return clientcmd.WriteToFile(rawConfig, tempFile)
}

func (l *Login) testKubernetes(tempFile string) (clientcmd.ClientConfig, error) {
	if err := l.selectContextAndWriteToFile(tempFile); err != nil {
		return nil, err
	}

	clientConfig, err := loadKubeConfig(tempFile)
	if err != nil {
		return nil, err
	}

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, nil
	}

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	info, err := client.Discovery().ServerVersion()
	if err != nil {
		return nil, err
	}

	if !is110OrGreater(*info) {
		return nil, fmt.Errorf("Found Kubernetes v%s.%s but v1.10 or newer is required", info.Major, info.Minor)
	}

	_, err = client.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rio-system",
		},
	})
	if err != nil && !errors2.IsConflict(err) {
		return nil, err
	}

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

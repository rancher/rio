package login

import (
	"io/ioutil"
	"os"

	"github.com/rancher/rio/cli/server"
	"github.com/rancher/rio/pkg/deploy/stack"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types/apis/rio.cattle.io/v1beta1"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
)

func InstallRioInK8s(config clientcmd.ClientConfig) error {
	ch, err := server.ConfigHome()
	if err != nil {
		return err
	}

	kubeConfig, err := ioutil.TempFile(ch, "tmp-")
	if err != nil {
		return err
	}
	if err := kubeConfig.Close(); err != nil {
		return err
	}
	defer os.Remove(kubeConfig.Name())

	rawConfig, err := config.RawConfig()
	if err != nil {
		return err
	}

	if err := clientcmd.WriteToFile(rawConfig, kubeConfig.Name()); err != nil {
		return err
	}

	old := os.Getenv("KUBECONFIG")
	os.Setenv("KUBECONFIG", kubeConfig.Name())
	defer os.Setenv("KUBECONFIG", old)

	service := &v1beta1.Service{}
	service.Name = "rio"
	service.Namespace = "rio-system"
	service.Spec.Image = settings.RioFullImage()
	service.Spec.Scale = 1
	service.Spec.ExposedPorts = []v1beta1.ExposedPort{
		{
			Name: "https",
			PortBinding: v1beta1.PortBinding{
				TargetPort: 5443,
				Port:       443,
			},
		},
	}
	service.Spec.GlobalPermissions = []v1beta1.Permission{
		{
			APIGroup: "*",
			Verbs:    []string{"*"},
			Resource: "*",
		},
	}

	logrus.Infof("Installing Rio")
	return stack.Deploy("rio-system", nil, nil, []*v1beta1.Service{service}, nil, nil)
}

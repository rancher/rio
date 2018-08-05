package login

import (
	"io/ioutil"
	"os"

	"github.com/rancher/rio/cli/pkg/up/questions"
	"github.com/rancher/rio/cli/server"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type Login struct {
	S_Server   string `desc:"Server to log into"`
	T_Token    string `desc:"Authentication token"`
	Kubeconfig string `desc:"Kubeconfig to use for existing Kubernetes setup"`
}

func (l *Login) useK8s() (bool, error) {
	if l.Kubeconfig != "" {
		return true, nil
	}

	if l.T_Token != "" || l.S_Server != "" {
		return false, nil
	}

	_, err := loadKubeConfig("")
	if err != nil {
		// purposefully mask error
		return false, nil
	}

	num, err := questions.PromptOptions("\n", 1,
		"[1] Connect to remote Rio server\n",
		"[2] Install Rio in existing Kubernetes\n",
	)
	return num == 1, err
}

func (l *Login) Run(app *cli.Context) (ex error) {
	defer func() {
		if ex == nil {
			logrus.Infof("Log in successful")
		}
	}()

	configHome, rioConf, k8sConf, err := server.Paths()
	if err != nil {
		return err
	}
	os.MkdirAll(configHome, 0700)

	f, err := ioutil.TempFile(configHome, "tmp-")
	if err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	defer os.Remove(f.Name())

	k8s, err := l.useK8s()
	if err != nil {
		return err
	}

	if k8s {
		err = l.k8s(f.Name())
	} else {
		err = l.remote(f.Name())
	}
	if err != nil {
		return err
	}

	_, err = server.SpaceClient(f.Name(), k8s)
	if err != nil {
		return err
	}

	if k8s {
		os.Remove(rioConf)
		return os.Rename(f.Name(), k8sConf)
	}

	os.Remove(k8sConf)
	return os.Rename(f.Name(), rioConf)
}

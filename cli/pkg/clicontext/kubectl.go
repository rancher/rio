package clicontext

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/rancher/wrangler/pkg/kubeconfig"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
)

func (c *Config) Kubectl(namespace, command string, args ...string) error {
	var execArgs []string
	if logrus.GetLevel() >= logrus.DebugLevel {
		execArgs = append(execArgs, "--v=9")
	}
	if namespace != "" {
		execArgs = append(execArgs, "-n", namespace)
	}
	if command != "" {
		execArgs = append(execArgs, command)
	}
	execArgs = append(execArgs, args...)

	if c.Kubeconfig == "" {
		loader := kubeconfig.GetInteractiveClientConfig(c.Kubeconfig)
		rawconfig, err := loader.RawConfig()
		if err != nil {
			return err
		}
		fp, err := ioutil.TempFile("", "kubeconfig-")
		if err != nil {
			return err
		}
		if err := fp.Close(); err != nil {
			return err
		}
		defer os.Remove(fp.Name())
		if err := clientcmd.WriteToFile(rawconfig, fp.Name()); err != nil {
			return err
		}
		c.Kubeconfig = fp.Name()
	}
	logrus.Debugf("kubectl %v, KUBECONFIG=%s", execArgs, c.Kubeconfig)
	cmd := exec.Command("kubectl", execArgs...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", c.Kubeconfig))
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

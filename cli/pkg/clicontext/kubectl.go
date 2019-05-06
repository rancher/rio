package clicontext

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

func (c *Config) KubectlCmd(namespace, command string, args ...string) (*exec.Cmd, error) {
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

	logrus.Debugf("kubectl %v, KUBECONFIG=%s", execArgs, c.Kubeconfig)
	cmd := exec.Command("kubectl", execArgs...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", c.Kubeconfig))
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd, nil
}

func (c *Config) Kubectl(namespace, command string, args ...string) error {
	cmd, err := c.KubectlCmd(namespace, command, args...)
	if err != nil {
		return err
	}
	return cmd.Run()
}

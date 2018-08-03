package containerd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

func Run() {
	args := []string{
		"containerd",
		"-a", "/run/rio/containerd.sock",
		"--state", "/run/rio/containerd",
	}

	if logrus.GetLevel() >= logrus.DebugLevel {
		args = append(args, "--verbose")
	}

	go func() {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Pdeathsig: syscall.SYS_KILL,
		}
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "containerd: %s\n", err)
		}
		os.Exit(1)
	}()

	time.Sleep(1 * time.Second)
}

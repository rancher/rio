package systemlogs

import (
	"bufio"
	"fmt"

	"github.com/docker/libcompose/cli/logger"
	"github.com/rancher/rio/cli/pkg/clicontext"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SystemLogs struct {
}

func (s SystemLogs) Run(ctx *clicontext.CLIContext) error {
	pods, err := ctx.Core.Pods(ctx.SystemNamespace).List(metav1.ListOptions{
		LabelSelector: "rio-controller=true",
	})
	if err != nil {
		return err
	}

	if len(pods.Items) == 0 {
		return fmt.Errorf("failed to find rio controller pod")
	}

	factory := logger.NewColorLoggerFactory()
	logger := factory.CreateContainerLogger("rio-controller")
	req := ctx.Core.Pods(pods.Items[0].Namespace).GetLogs(pods.Items[0].Name, &v1.PodLogOptions{
		Follow: true,
	})
	reader, err := req.Stream()
	if err != nil {
		return err
	}
	defer reader.Close()

	sc := bufio.NewScanner(reader)
	for sc.Scan() {
		logger.Out(append(sc.Bytes(), []byte("\n")...))
	}

	return nil
}

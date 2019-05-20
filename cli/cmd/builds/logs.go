package builds

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/docker/libcompose/cli/logger"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	"github.com/rancher/wrangler/pkg/kv"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Logs struct {
}

func (l *Logs) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return fmt.Errorf("at least one argument is required")
	}

	serviceName, revision := kv.Split(ctx.CLI.Args()[0], ":")
	namespace, name := stack.NamespaceAndName(ctx, serviceName)
	pods, err := ctx.Core.Pods("").List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("service-name=%s, service-namespace=%s", name, namespace),
	})
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		if !strings.Contains(pod.Labels["build.knative.dev/buildName"], revision[0:13]) {
			continue
		}
		factory := logger.NewColorLoggerFactory()
		for _, container := range append(pod.Spec.InitContainers, pod.Spec.Containers...) {
			logger := factory.CreateContainerLogger(container.Name)
			req := ctx.Core.Pods(pod.Namespace).GetLogs(pod.Name, &v1.PodLogOptions{
				Follow:    true,
				Container: container.Name,
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
		}

	}

	return nil
}

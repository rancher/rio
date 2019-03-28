package logs

import (
	"bufio"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/docker/libcompose/cli/logger"
	"github.com/rancher/rio/cli/cmd/ps"
	"github.com/rancher/rio/cli/pkg/clicontext"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/apimachinery/pkg/runtime/schema"
	_ "k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type Logs struct {
	F_Follow    bool   `desc:"Follow log output"`
	S_Since     string `desc:"Logs since a certain time, either duration (5s, 2m, 3h) or RFC3339"`
	P_Previous  bool   `desc:"Print the logs for the previous instance of the container in a pod if it exists"`
	C_Container string `desc:"Print the logs of a specific container"`
	N_Tail      int    `desc:"Number of recent lines of logs to print, -1 for all" default:"200"`
	A_All       bool   `desc:"Include hidden or systems logs when logging"`
}

func (l *Logs) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) == 0 {
		return fmt.Errorf("at least one argument is required: CONTAINER_OR_SERVICE")
	}

	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}

	restClient, err := cluster.RestClient()
	if err != nil {
		return err
	}

	pds, err := ps.ListPods(ctx, l.A_All, ctx.CLI.Args()...)
	if err != nil {
		return err
	}

	if len(pds) == 0 {
		return fmt.Errorf("failed to find container for %v, container \"%s\"", ctx.CLI.Args(), l.C_Container)
	}

	factory := logger.NewColorLoggerFactory()
	for _, pd := range pds {
		for _, container := range pd.Containers {
			if l.C_Container == "" || l.C_Container == container.Name {
				go func(c string) {
					l.logContainer(pd.Pod, c, restClient, factory)
				}(container.Name)
			}
		}
	}
	<-ctx.Ctx.Done()

	return nil
}

func (l *Logs) logContainer(pod *v1.Pod, containerName string, restClient rest.Interface, factory *logger.ColorLoggerFactory) error {
	cn := fmt.Sprintf("%s/%s", pod.Name, containerName)
	logger := factory.CreateContainerLogger(cn)
	podLogOption := v1.PodLogOptions{
		Container: containerName,
		Follow:    l.F_Follow,
	}
	if l.S_Since != "" {
		t, err := time.Parse(time.RFC3339, l.S_Since)
		if err == nil {
			newtime := metav1.NewTime(t)
			podLogOption.SinceTime = &newtime
		} else {
			du, err := time.ParseDuration(l.S_Since)
			if err == nil {
				ss := int64(du.Round(time.Second).Seconds())
				podLogOption.SinceSeconds = &ss
			}
		}
	}

	scheme.Scheme.AddUnversionedTypes(v1.SchemeGroupVersion, &v1.PodLogOptions{})
	req := restClient.Get().
		Prefix("api", "", "v1").
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("log").
		VersionedParams(&podLogOption, runtime.NewParameterCodec(scheme.Scheme))
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

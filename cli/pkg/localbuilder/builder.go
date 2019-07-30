package localbuilder

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rancher/rio/cli/pkg/localbuilder/containerd"
	"github.com/rancher/rio/cli/pkg/localbuilder/docker"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

var (
	dockerSocket      = "/run/docker.sock"
	containerdSocket  = "/run/k3s/containerd/containerd.sock"
	buildkitNamespace = "default"

	maxBuildThread = int64(10)

	base = 32768
	end  = 61000
)

type LocalBuilder interface {
	Build(ctx context.Context, specs map[string]riov1.ImageBuild, parallel bool, namespace string) (map[string]string, error)
}

type RuntimeBuilder interface {
	Build(ctx context.Context, spec riov1.ImageBuild) (string, error)
}

func NewLocalBuilder(ctx context.Context, apply apply.Apply, k8s *kubernetes.Clientset) (LocalBuilder, error) {
	runtime := containerRuntime(k8s)
	builder := &localBuilder{
		runtime:      runtime,
		k8s:          k8s,
		apply:        apply,
		buildkitPort: generateRandomPort(),
		socketPort:   generateRandomPort(),
	}

	var err error
	if runtime == "docker" {
		builder.socketAddress = dockerSocket
		builder.runtimeBuilder = docker.NewDockerBuilder()
	} else if runtime == "containerd" {
		builder.runtimeBuilder, err = containerd.NewContainerdBuilder(builder.socketPort, builder.buildkitPort)
		if err != nil {
			return nil, err
		}
		builder.socketAddress = containerdSocket
	} else {
		return nil, fmt.Errorf("runtime %s is not supported for local builder", runtime)
	}

	return builder, builder.setup(ctx)
}

type localBuilder struct {
	socketAddress  string
	apply          apply.Apply
	k8s            *kubernetes.Clientset
	runtime        string
	runtimeBuilder RuntimeBuilder
	buildkitPort   string
	socketPort     string
}

func (l localBuilder) Build(ctx context.Context, specs map[string]riov1.ImageBuild, parallel bool, namespace string) (map[string]string, error) {
	if !parallel {
		maxBuildThread = 1
	}

	result := make(map[string]string)
	m := sync.Map{}
	errg, _ := errgroup.WithContext(ctx)
	s := semaphore.NewWeighted(maxBuildThread)

	specs = setupBuildConfig(specs, namespace)
	for name := range specs {
		n := name
		if err := s.Acquire(ctx, 1); err != nil {
			return nil, err
		}
		errg.Go(func() error {
			image, err := l.buildSingle(ctx, specs[n])
			if err != nil {
				return err
			}

			m.LoadOrStore(n, image)
			s.Release(1)
			return nil
		})
	}
	if err := errg.Wait(); err != nil {
		return result, err
	}

	m.Range(func(key, value interface{}) bool {
		result[key.(string)] = value.(string)
		return true
	})
	return result, nil
}

func (l localBuilder) setup(ctx context.Context) error {
	if err := l.setupStack(); err != nil {
		return err
	}

	return l.setupPortforwarding(ctx)
}

func (l localBuilder) setupStack() error {
	return stack.NewSystemStack(l.apply, buildkitNamespace, "buildkit-local").Deploy(map[string]string{
		"RUNTIME":        l.runtime,
		"SOCKET_ADDRESS": l.socketAddress,
	})
}

func (l localBuilder) setupPortforwarding(ctx context.Context) error {
	newctx, cancel := context.WithCancel(ctx)

	var deployReady bool
	var socatPod *v1.Pod
	wait.JitterUntil(func() {
		if !deployReady {
			if deploy, err := l.k8s.AppsV1().Deployments(buildkitNamespace).Get("buildkit", metav1.GetOptions{}); err == nil {
				if isReady(&deploy.Status) {
					deployReady = true
				}
				logrus.Debug("Waiting for buildkitd deploy to be ready")
			}
		}

		if p, err := l.k8s.CoreV1().Pods(buildkitNamespace).Get("socat-socket", metav1.GetOptions{}); err == nil {
			if p.Status.Phase == v1.PodRunning {
				socatPod = p
				cancel()
				return
			}
			logrus.Debug("Waiting for socat-socket pod to be running")
		}
	}, 100*time.Millisecond, 1.5, false, newctx.Done())

	buildkitdPod, err := findPod(l.k8s, buildkitNamespace, "app=buildkitd-dev")
	if err != nil {
		return err
	}

	go func() {
		if err := portForward(buildkitdPod, l.k8s, l.buildkitPort, "8080", chanWrapper(ctx.Done())); err != nil {
			logrus.Error(err)
		}
	}()

	go func() {
		if err := portForward(socatPod, l.k8s, l.socketPort, "9091", chanWrapper(ctx.Done())); err != nil {
			logrus.Error(err)
		}
	}()

	return nil
}

func chanWrapper(input <-chan struct{}) chan struct{} {
	output := make(chan struct{}, 1)
	go func() {
		select {
		case s := <-input:
			output <- s
		}
	}()
	return output
}

func setupBuildConfig(specs map[string]riov1.ImageBuild, namespace string) map[string]riov1.ImageBuild {
	r := map[string]riov1.ImageBuild{}
	for name, config := range specs {
		if config.DockerFile == "" {
			config.DockerFile = "Dockerfile"
		}
		if config.BuildContext == "" {
			config.BuildContext = "."
		}
		if config.DockerFilePath == "" {
			config.DockerFilePath = config.BuildContext
		}
		if config.BuildImageName == "" {
			config.BuildImageName = fmt.Sprintf("%s/%s", namespace, name)
		}
		if config.PushRegistry == "" {
			config.PushRegistry = "docker.io"
		}
		r[name] = config
	}
	return r
}

func (l localBuilder) buildSingle(ctx context.Context, spec riov1.ImageBuild) (string, error) {
	newctx, cancel := context.WithCancel(ctx)
	var image string
	var err error
	wait.JitterUntil(func() {
		image, err = l.runtimeBuilder.Build(newctx, spec)
		if err == nil {
			cancel()
		} else if err != nil && !strings.Contains(err.Error(), "connect: connection refused") {
			logrus.Fatal(err)
		}
	}, time.Second, 1.5, false, newctx.Done())
	return image, err
}

func generateRandomPort() string {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	for {
		port := base + r1.Intn(end-base+1)
		ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
		if err != nil {
			continue
		}
		ln.Close()
		return strconv.Itoa(port)
	}
}

func containerRuntime(k8s *kubernetes.Clientset) string {
	nodes, err := k8s.CoreV1().Nodes().List(metav1.ListOptions{})
	if err == nil && len(nodes.Items) == 1 && strings.Contains(nodes.Items[0].Status.NodeInfo.ContainerRuntimeVersion, "containerd") {
		return "containerd"
	}
	return "docker"
}

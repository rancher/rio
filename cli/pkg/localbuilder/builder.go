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

	"github.com/rancher/rio/cli/pkg/localbuilder/runc"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
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
	maxBuildThread = int64(10)

	base = 32768
	end  = 61000
)

type LocalBuilder interface {
	Build(ctx context.Context, specs map[stack.ContainerBuildKey]riov1.ImageBuildSpec, parallel bool, namespace string) (map[stack.ContainerBuildKey]string, error)
}

type RuntimeBuilder interface {
	Build(ctx context.Context, spec riov1.ImageBuildSpec) (string, error)
}

func NewLocalBuilder(ctx context.Context, systemNamespace string, apply apply.Apply, k8s *kubernetes.Clientset) (LocalBuilder, error) {
	builder := &localBuilder{
		k8s:             k8s,
		apply:           apply,
		buildkitPort:    generateRandomPort(),
		systemNamespace: systemNamespace,
	}

	builder.runtimeBuilder = runc.NewRuncBuilder(builder.buildkitPort)

	return builder, builder.setup(ctx)
}

type localBuilder struct {
	apply           apply.Apply
	k8s             *kubernetes.Clientset
	runtimeBuilder  RuntimeBuilder
	buildkitPort    string
	systemNamespace string
}

func (l localBuilder) Build(ctx context.Context, specs map[stack.ContainerBuildKey]riov1.ImageBuildSpec, parallel bool, namespace string) (map[stack.ContainerBuildKey]string, error) {
	if !parallel {
		maxBuildThread = 1
	}

	result := make(map[stack.ContainerBuildKey]string)
	m := sync.Map{}
	errg, _ := errgroup.WithContext(ctx)
	s := semaphore.NewWeighted(maxBuildThread)

	specs = l.setupBuildConfig(specs, namespace)
	for name := range specs {
		n := name
		if err := s.Acquire(ctx, 1); err != nil {
			return nil, err
		}
		errg.Go(func() error {
			logrus.Infof("Building service %s", n)
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
		result[key.(stack.ContainerBuildKey)] = value.(string)
		return true
	})
	return result, nil
}

func (l localBuilder) setup(ctx context.Context) error {
	return l.setupPortforwarding(ctx)
}

func (l localBuilder) setupPortforwarding(ctx context.Context) error {
	go func() {
		pods, err := l.k8s.CoreV1().Pods(l.systemNamespace).List(metav1.ListOptions{
			LabelSelector: fmt.Sprintf("app=%s", constants.BuildkitdService),
		})
		if err != nil {
			logrus.Error(err)
		}
		var pod v1.Pod
		for _, p := range pods.Items {
			if p.Status.Phase == v1.PodRunning {
				pod = p
				break
			}
		}
		if err := PortForward(l.k8s, l.buildkitPort, "8080", pod, false, ChanWrapper(ctx.Done())); err != nil {
			logrus.Error(err)
		}
	}()

	return nil
}

func ChanWrapper(input <-chan struct{}) chan struct{} {
	output := make(chan struct{}, 1)
	go func() {
		select {
		case s := <-input:
			output <- s
		}
	}()
	return output
}

func (l localBuilder) setupBuildConfig(specs map[stack.ContainerBuildKey]riov1.ImageBuildSpec, namespace string) map[stack.ContainerBuildKey]riov1.ImageBuildSpec {
	r := map[stack.ContainerBuildKey]riov1.ImageBuildSpec{}
	for name, config := range specs {
		if config.Dockerfile == "" {
			config.Dockerfile = "Dockerfile"
		}
		if config.Context == "" {
			config.Context = "."
		}
		if config.ImageName == "" {
			config.ImageName = fmt.Sprintf("%s/%s", namespace, name)
		}
		if config.PushRegistry == "" {
			config.PushRegistry = constants.RegistryService
		}
		r[name] = config
	}
	return r
}

func (l localBuilder) buildSingle(ctx context.Context, spec riov1.ImageBuildSpec) (string, error) {
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

package docker

import (
	"context"
	"fmt"

	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
)

// Todo: minikube doesn't support buildkit, to be implemented as the current hack is not good
type Builder struct {
}

func NewDockerBuilder() Builder {
	return Builder{}
}

func (b Builder) Build(ctx context.Context, spec riov1.ImageBuild) (string, error) {
	return "", fmt.Errorf("docker runtime is not supported currently")
}

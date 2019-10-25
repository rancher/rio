package runc

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/containerd/console"
	"github.com/containerd/containerd"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/cmd/buildctl/build"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/util/progress/progressui"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"golang.org/x/sync/errgroup"
)

type Builder struct {
	containerdClient *containerd.Client
	buildkitPort     string
}

func NewRuncBuilder(buildkitPort string) *Builder {
	return &Builder{
		buildkitPort: buildkitPort,
	}
}

func (l Builder) Build(ctx context.Context, spec riov1.ImageBuildSpec) (string, error) {
	return l.build(ctx, spec)
}

func (l Builder) build(ctx context.Context, spec riov1.ImageBuildSpec) (string, error) {
	buildkitClient, solveOpt, err := initializeClient(ctx, spec, l.buildkitPort)
	if err != nil {
		return "", nil
	}

	image := fmt.Sprintf("%s/%s", spec.PushRegistry, spec.ImageName)
	exportEntry := client.ExportEntry{
		Type:  "image",
		Attrs: map[string]string{},
	}
	exportEntry.Attrs["name"] = image
	exportEntry.Attrs["push"] = "true"

	if strings.HasPrefix(spec.PushRegistry, constants.BuildkitdService) {
		exportEntry.Attrs["registry.insecure"] = "true"
	}
	solveOpt.Exports = []client.ExportEntry{
		exportEntry,
	}

	ch := make(chan *client.SolveStatus)
	eg, _ := errgroup.WithContext(ctx)
	displayCh := ch

	var digest string
	eg.Go(func() error {
		resp, err := buildkitClient.Solve(ctx, nil, solveOpt, ch)
		if err != nil {
			return err
		}
		for k, v := range resp.ExporterResponse {
			if k == "containerimage.digest" {
				digest = v
			}
		}

		if !strings.ContainsAny(spec.ImageName, ":@") {
			image = fmt.Sprintf("%s@%s", image, digest)
		}
		return nil
	})

	eg.Go(func() error {
		var c console.Console
		cf, _ := console.ConsoleFromFile(os.Stderr)
		c = cf
		return progressui.DisplaySolveStatus(context.TODO(), "", c, os.Stderr, displayCh)
	})
	if err := eg.Wait(); err != nil {
		return "", err
	}

	return image, nil
}

func initializeClient(ctx context.Context, buildSpec riov1.ImageBuildSpec, port string) (*client.Client, client.SolveOpt, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	attachable := []session.Attachable{authprovider.NewDockerAuthProvider(os.Stderr)}

	buildkitClient, err := client.New(ctx, fmt.Sprintf("tcp://localhost:%s", port))
	if err != nil {
		return nil, client.SolveOpt{}, err
	}

	solveOpt := client.SolveOpt{
		Frontend: "dockerfile.v0",
		FrontendAttrs: map[string]string{
			"filename": buildSpec.Dockerfile,
		},
		LocalDirs: map[string]string{
			"context":    buildSpec.Context,
			"dockerfile": buildSpec.Context,
		},
		Session: attachable,
	}

	var buildArgs []string
	for _, arg := range buildSpec.Args {
		buildArgs = append(buildArgs, fmt.Sprintf("build-arg:%s", arg))
	}
	solveOpt.FrontendAttrs, err = build.ParseOpt(buildArgs, nil)
	if err != nil {
		return nil, client.SolveOpt{}, err
	}
	solveOpt.FrontendAttrs["filename"] = buildSpec.Dockerfile
	if buildSpec.NoCache {
		solveOpt.FrontendAttrs["no-cache"] = ""
	}

	return buildkitClient, solveOpt, nil
}

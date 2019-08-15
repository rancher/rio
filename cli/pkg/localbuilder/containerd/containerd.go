package containerd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/containerd/console"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/namespaces"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/cmd/buildctl/build"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/util/progress/progressui"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Builder struct {
	containerdClient *containerd.Client
	buildkitPort     string
}

func NewContainerdBuilder(socketPort, buildkitPort string) (*Builder, error) {
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%s", socketPort), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	containerdClient, err := containerd.NewWithConn(conn)
	return &Builder{
		containerdClient: containerdClient,
		buildkitPort:     buildkitPort,
	}, err
}

func (l Builder) Build(ctx context.Context, spec riov1.ImageBuild) (string, error) {
	tag, err := l.build(ctx, spec)
	if err != nil {
		return "", err
	}
	return tag, l.tag(ctx, spec, tag)
}

func (l Builder) tag(ctx context.Context, spec riov1.ImageBuild, tag string) error {
	if spec.Push {
		return nil
	}
	if !strings.ContainsAny(spec.BuildImageName, ":@") {
		image := fmt.Sprintf("%s/%s", spec.PushRegistry, spec.BuildImageName)
		if err := l.tagContainerd(ctx, image, tag); err != nil {
			return err
		}
	}
	return nil
}

func (l Builder) tagContainerd(ctx context.Context, original, tag string) error {
	md := metadata.New(map[string]string{namespaces.GRPCHeader: "k8s.io"})
	ctx = metadata.NewOutgoingContext(ctx, md)
	originalImage, err := l.containerdClient.ImageService().Get(ctx, original)
	if err != nil {
		return err
	}

	img := images.Image{
		Name:      tag,
		Target:    originalImage.Target,
		CreatedAt: time.Now(),
	}

	if _, err := l.containerdClient.ImageService().Create(ctx, img); err != nil && !strings.Contains(err.Error(), "already exists") {
		return err
	}

	return nil
}

func (l Builder) build(ctx context.Context, spec riov1.ImageBuild) (string, error) {
	buildkitClient, solveOpt, err := initializeClient(ctx, spec, l.buildkitPort)
	if err != nil {
		return "", nil
	}

	image := fmt.Sprintf("%s/%s", spec.PushRegistry, spec.BuildImageName)
	exportEntry := client.ExportEntry{
		Type:  "image",
		Attrs: map[string]string{},
	}
	exportEntry.Attrs["name"] = image
	if spec.Push {
		exportEntry.Attrs["push"] = "true"
	}
	if strings.HasPrefix(spec.PushRegistry, "registry.") {
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

		if !strings.ContainsAny(spec.BuildImageName, ":@") {
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

func initializeClient(ctx context.Context, buildSpec riov1.ImageBuild, port string) (*client.Client, client.SolveOpt, error) {
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
			"filename": buildSpec.DockerFile,
		},
		LocalDirs: map[string]string{
			"context":    buildSpec.BuildContext,
			"dockerfile": buildSpec.DockerFilePath,
		},
		Session: attachable,
	}

	var buildArgs []string
	for _, arg := range buildSpec.BuildArgs {
		buildArgs = append(buildArgs, fmt.Sprintf("build-arg:%s", arg))
	}
	solveOpt.FrontendAttrs, err = build.ParseOpt(buildArgs, nil)
	if err != nil {
		return nil, client.SolveOpt{}, err
	}
	solveOpt.FrontendAttrs["filename"] = buildSpec.DockerFile
	if buildSpec.NoCache {
		solveOpt.FrontendAttrs["no-cache"] = ""
	}

	return buildkitClient, solveOpt, nil
}

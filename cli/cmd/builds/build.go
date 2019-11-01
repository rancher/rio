package builds

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/rancher/rio/cli/cmd/up/pkg"
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stack"
	"github.com/urfave/cli"
	"sigs.k8s.io/yaml"
)

func Builds(app *cli.App) cli.Command {
	ls := builder.Command(&History{},
		"List Build History",
		app.Name+" build history [OPTIONS]",
		"")
	build := builder.Command(&Build{},
		"Build docker image using buildkitd",
		app.Name+" build [OPTIONS] PATH",
		"")
	build.Subcommands = []cli.Command{
		ls,
	}
	return build
}

type Build struct {
	F_File   string   `desc:"Name of the file to look for build, support both Riofile and Dockerfile"`
	T_Tag    string   `desc:"Name and optionally a tag in the 'name:tag' format"`
	BuildArg []string `desc:"Set build-time variables"`
	NoCache  bool     `desc:"Do not use cache when building the image"`
}

func (b *Build) Run(ctx *clicontext.CLIContext) error {
	content, err := pkg.LoadRiofile(b.F_File)
	if err != nil {
		return err
	}

	answers, err := pkg.LoadAnswer("")
	if err != nil {
		return err
	}

	deployStack := stack.NewStack(content, answers)
	imageBuilds, err := deployStack.GetImageBuilds()
	if err != nil {
		return err
	}

	if err := b.mergeBuildParameters(imageBuilds, ctx); err != nil {
		return err
	}

	images, err := pkg.Build(imageBuilds, ctx, false)
	if err != nil {
		return err
	}
	for _, im := range images {
		fmt.Println(im)
	}
	return nil
}

func (b *Build) mergeBuildParameters(buildSpecs map[stack.ContainerBuildKey]riov1.ImageBuildSpec, ctx *clicontext.CLIContext) error {
	useDockerfile := false
	if b.F_File != "" {
		content, err := ioutil.ReadFile(b.F_File)
		if err != nil {
			return err
		}
		var r map[interface{}]interface{}
		if err := yaml.Unmarshal(content, r); err == nil {
			useDockerfile = true
		}
	}

	for k, buildSpec := range buildSpecs {
		if len(ctx.CLI.Args()) == 1 {
			buildSpec.Context = filepath.Join(buildSpec.Context, ctx.CLI.Args()[0])
		}

		if b.T_Tag != "" {
			buildSpec.ImageName = fmt.Sprintf("%s/%s", ctx.GetSetNamespace(), b.T_Tag)
		}

		buildSpec.NoCache = b.NoCache

		if useDockerfile {
			rel, err := filepath.Rel(buildSpec.Context, b.F_File)
			if err != nil {
				return err
			}
			buildSpec.Dockerfile = rel
		}

		if len(b.BuildArg) > 0 {
			buildSpec.Args = b.BuildArg
		}

		buildSpecs[k] = buildSpec
	}
	return nil
}

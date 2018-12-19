package feature

import (
	"fmt"

	"github.com/rancher/norman/types"
	"github.com/rancher/rio/cli/cmd/project"
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/pkg/up/questions"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/types/client/project/v1"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	"github.com/urfave/cli"
)

func Feature(app *cli.App) cli.Command {
	ls := builder.Command(&Ls{},
		"List features",
		app.Name+" feature ls",
		"")
	return cli.Command{
		Name:      "features",
		ShortName: "feature",
		Usage:     "Enable, Disable and List features",
		Action:    clicontext.DefaultAction(ls.Action),
		Flags:     table.WriterFlags(),
		Category:  "SUB COMMANDS",
		Subcommands: []cli.Command{
			ls,
			builder.Command(&Enable{},
				"Enable a feature",
				app.Name+" feature enable $FEATURE_NAME", ""),
			builder.Command(&Disable{},
				"Disable a feature",
				app.Name+" feature disable $FEATURE_NAME", ""),
		},
	}
}

type Ls struct{}

type Data struct {
	ID      string
	Feature client.Feature
}

func (l *Ls) Run(ctx *clicontext.CLIContext) error {
	writer := table.NewWriter([][]string{
		{"NAME", "Feature.Name"},
		{"ENABLED", "{{.Feature.Enabled | boolToStar}}"},
		{"DESCRIPTION", "Feature.Description"},
	}, ctx)
	defer writer.Close()

	writer.AddFormatFunc("boolToStar", project.BoolToStar)
	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}
	spaceClient, err := cluster.Client()
	if err != nil {
		return err
	}
	featuresCollection, err := spaceClient.Feature.List(&types.ListOpts{})
	if err != nil {
		return err
	}

	for _, feature := range featuresCollection.Data {
		writer.Write(Data{
			ID:      feature.ID,
			Feature: feature,
		})
	}
	return writer.Err()
}

type Disable struct{}

func (d *Disable) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) != 1 {
		return fmt.Errorf("feature name is required")
	}

	ctx.ProjectName = settings.RioSystemNamespace
	return flipEnableFlag(ctx, ctx.CLI.Args()[0], false)
}

type Enable struct{}

func (e *Enable) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) != 1 {
		return fmt.Errorf("feature name is required")
	}
	ctx.ProjectName = settings.RioSystemNamespace
	return flipEnableFlag(ctx, ctx.CLI.Args()[0], true)
}

func flipEnableFlag(ctx *clicontext.CLIContext, featureName string, enable bool) error {
	cluster, err := ctx.Cluster()
	if err != nil {
		return err
	}
	spaceClient, err := cluster.Client()
	if err != nil {
		return err
	}
	feature, err := spaceClient.Feature.ByID(featureName)
	if err != nil {
		return err
	}
	feature.Enabled = enable
	if enable {
		qs, err := questions.NewQuestions(toQuestions(feature.Questions), feature.Answers, true)
		if err != nil {
			return err
		}
		answers, err := qs.Ask()
		if err != nil {
			return err
		}
		feature.Answers = answers
	}
	_, err = spaceClient.Feature.Replace(feature)
	if err != nil {
		return err
	}
	return nil
}

func toQuestions(qs []client.Question) []v3.Question {
	var r []v3.Question
	for _, q := range qs {
		r = append(r, v3.Question{
			Variable:    q.Variable,
			Description: q.Description,
			Label:       q.Label,
			Options:     q.Options,
			Default:     q.Default,
		})
	}
	return r
}

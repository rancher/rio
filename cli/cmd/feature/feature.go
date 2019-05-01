package feature

import (
	"fmt"

	"github.com/rancher/rio/cli/cmd/up"
	"github.com/rancher/rio/cli/pkg/builder"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/lookup"
	"github.com/rancher/rio/cli/pkg/table"
	"github.com/rancher/rio/cli/pkg/tables"
	clitypes "github.com/rancher/rio/cli/pkg/types"
	"github.com/rancher/rio/cli/pkg/up/questions"
	projectv1 "github.com/rancher/rio/pkg/apis/project.rio.cattle.io/v1"
	"github.com/urfave/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Feature projectv1.Feature
}

func (l *Ls) Run(ctx *clicontext.CLIContext) error {
	features, err := ctx.List(clitypes.FeatureType)
	if err != nil {
		return err
	}
	writer := tables.NewFeature(ctx)
	return writer.Write(features)
}

type Disable struct{}

func (d *Disable) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) != 1 {
		return fmt.Errorf("feature name is required")
	}

	resource, err := lookup.Lookup(ctx, ctx.CLI.Args()[0], clitypes.FeatureType)
	if err != nil {
		return err
	}

	return flipEnableFlag(ctx, resource.Name, nil, false)
}

type Enable struct {
	A_Answers string `desc:"Answer file in with key/value pairs in yaml or json"`
}

func (e *Enable) Run(ctx *clicontext.CLIContext) error {
	if len(ctx.CLI.Args()) != 1 {
		return fmt.Errorf("feature name is required")
	}
	resource, err := lookup.Lookup(ctx, ctx.CLI.Args()[0], clitypes.FeatureType)
	if err != nil {
		return err
	}
	answers, err := up.ReadAnswers(e.A_Answers)
	if err != nil {
		return fmt.Errorf("failed to parse answer file [%s]: %v", e.A_Answers, err)
	}
	return flipEnableFlag(ctx, resource.Name, answers, true)
}

func flipEnableFlag(ctx *clicontext.CLIContext, featureName string, answers map[string]string, enable bool) error {
	feature, err := ctx.Project.Features(ctx.Namespace).Get(featureName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	feature.Status.EnableOverride = &enable

	if enable {
		if len(answers) == 0 {
			qs, err := questions.NewQuestions(feature.Spec.Questions, feature.Spec.Answers, true)
			if err != nil {
				return err
			}
			answers, err := qs.Ask()
			if err != nil {
				return err
			}
			feature.Spec.Answers = answers
		} else {
			if feature.Spec.Answers == nil {
				feature.Spec.Answers = map[string]string{}
			}
			for k, v := range answers {
				feature.Spec.Answers[k] = v
			}
		}
	}
	return ctx.UpdateObject(feature)
}

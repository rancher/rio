package render

import (
	"bytes"
	"fmt"

	"sigs.k8s.io/yaml"

	"github.com/rancher/rio/cli/pkg/table"

	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/cmd/apply"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/up/questions"
	"github.com/rancher/rio/modules/service/controllers/service/populate/servicelabels"
	"github.com/rancher/rio/pkg/systemstack"
)

type Render struct {
	N_Namepsace string `desc:"Namespace to apply" default:"default"`
	F_File      string `desc:"the path to the rio file to apply"`
	A_Answers   string `desc:"Answer file in with key/value pairs in yaml or json"`
	Prompt      bool   `desc:"Re-ask all questions if answer is not found in environment variables"`
}

func (r *Render) Run(ctx *clicontext.CLIContext) error {
	content, err := util.ReadFile(r.F_File)
	if err != nil {
		return errors.Wrapf(err, "reading %s", r.F_File)
	}

	answers, err := apply.ReadAnswers(r.A_Answers)
	if err != nil {
		return fmt.Errorf("failed to parse answer file [%s]: %v", r.A_Answers, err)
	}

	stack := systemstack.NewStack(ctx.Apply, r.N_Namepsace, r.N_Namepsace, true)
	stack.WithContent(content)

	qs, err := stack.Questions()
	if err != nil {
		return err
	}

	newQuestions, err := questions.NewQuestions(qs, answers, r.Prompt)
	if err != nil {
		return err
	}
	newAnswers, err := newQuestions.Ask()
	if err != nil {
		return err
	}
	mergedAnswers := servicelabels.Merge(answers, newAnswers)

	objs, err := stack.Objects(mergedAnswers)
	if err != nil {
		return err
	}
	buffer := &bytes.Buffer{}
	for _, obj := range objs {
		data, err := table.FormatJSON(obj)
		if err != nil {
			return err
		}
		converted, err := yaml.JSONToYAML([]byte(data))
		if err != nil {
			return err
		}
		buffer.Write(converted)
		buffer.WriteString("---\n")
	}
	fmt.Println(buffer.String())
	return nil
}

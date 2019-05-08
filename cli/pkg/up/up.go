package up

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/up/questions"
	"github.com/rancher/rio/modules/service/controllers/service/populate/servicelabels"
	"github.com/rancher/rio/pkg/systemstack"
)

func Run(ctx *clicontext.CLIContext, content []byte, namespace string, answers map[string]string, prompt bool) error {
	stack := systemstack.NewStack(ctx.Apply, namespace, namespace, true)
	stack.WithContent(content)

	qs, err := stack.Questions()
	if err != nil {
		return err
	}

	newQuestions, err := questions.NewQuestions(qs, answers, prompt)
	if err != nil {
		return err
	}
	newAnswers, err := newQuestions.Ask()
	if err != nil {
		return err
	}
	mergedAnswers := servicelabels.Merge(answers, newAnswers)

	return stack.Deploy(mergedAnswers)
}

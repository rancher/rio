package up

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/docker/docker/pkg/symlink"
	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/up/questions"
	"github.com/rancher/rio/pkg/stackfile"
	"github.com/rancher/rio/pkg/template"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func readFileInPath(relativePath, file string) ([]byte, error) {
	if strings.HasPrefix(relativePath, "http") {
		base, err := url.Parse(relativePath)
		if err != nil {
			return nil, err
		}
		ref, err := url.Parse(file)
		if err != nil {
			return nil, err
		}
		resolved := base.ResolveReference(ref)
		if err != nil {
			return nil, err
		}
		resp, err := http.Get(resolved.String())
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)
	}
	f, err := symlink.FollowSymlinkInScope(file, relativePath)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadFile(f)
}

func readFiles(relativePath string, files []string, template *stackfile.StackFile, promptReplaceFile bool) error {
	for _, file := range files {
		existingContent, exists := template.AdditionalFiles[file]
		content, err := readFileInPath(relativePath, file)
		if err != nil {
			if exists {
				continue
			}
			return fmt.Errorf("failed to read file %s: %v", file, err)
		}

		if exists && bytes.Compare(existingContent, content) != 0 {
			yn := true
			if promptReplaceFile {
				yn, err = questions.PromptBool(fmt.Sprintf("The contents of %s have changed, do you want to update", file), false)
				if err != nil {
					return errors.Wrap(err, "failed to ask question")
				}
			}

			if yn {
				template.AdditionalFiles[file] = content
			}
		} else if !exists {
			template.AdditionalFiles[file] = content
		}
	}

	return nil
}

func Run(ctx *clicontext.CLIContext, content []byte, stackName string, promptReplaceFile, prompt bool, answers map[string]string, file string) error {
	stack, err := ctx.Rio.Stacks(ctx.Namespace).Get(stackName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	template, err := stackfile.FromStack(stack)
	if err != nil {
		return err
	}
	template.Template.Content = content

	files, err := template.RequiredFiles()
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	relativePath := cwd
	if strings.HasPrefix(file, "http") {
		relativePath = file
	}
	if err := readFiles(relativePath, files, template, promptReplaceFile); err != nil {
		return err
	}

	if err := template.Validate(); err != nil {
		return fmt.Errorf("failed to parse template. If you are using go templating the template must execute with no values: %v", err)
	}

	mergeAnswers(&template.Template, answers)

	if err := template.PopulateAnswersFromEnv(); err != nil {
		return err
	}

	if err := populateAnswersFromQuestions(&template.Template, prompt); err != nil {
		return err
	}

	stack.Spec = template.ToStackSpec()
	return ctx.UpdateObject(stack)
}

func mergeAnswers(template *template.Template, answers map[string]string) {
	for k, v := range answers {
		template.Answers[k] = v
	}
}

func populateAnswersFromQuestions(template *template.Template, forcePrompt bool) error {
	qs, err := questions.NewQuestions(template.Questions, template.Answers, forcePrompt)
	if err != nil {
		return err
	}

	answers, err := qs.Ask()
	if err != nil {
		return err
	}

	template.Answers = answers
	return nil
}

package up

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/docker/docker/pkg/symlink"
	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/up/questions"
	template2 "github.com/rancher/rio/pkg/template"
)

func readFileInPath(cwd, file, filePath string) ([]byte, error) {
	if strings.HasPrefix(filePath, "http") {
		base, err := url.Parse(filePath)
		if err != nil {
			return nil, err
		}
		ref, err := url.Parse(filepath.Join(file))
		if err != nil {
			return nil, err
		}
		resolved := base.ResolveReference(ref)
		if err != nil {
			return nil, err
		}
		logrus.Infof(resolved.String())
		resp, err := http.Get(resolved.String())
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)
	}
	f, err := symlink.FollowSymlinkInScope(file, cwd)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadFile(f)
}

func readFiles(filePath string, files []string, template *template2.Template, promptReplaceFile bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "getwd")
	}

	for _, file := range files {
		existingContent, exists := template.AdditionalFiles[file]
		content, err := readFileInPath(cwd, file, filePath)
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

func Run(ctx *clicontext.CLIContext, content []byte, stackID string, promptReplaceFile, prompt bool, answers map[string]string, file string) error {
	wc, err := ctx.WorkspaceClient()
	if err != nil {
		return err
	}

	stack, err := wc.Stack.ByID(stackID)
	if err != nil {
		return err
	}

	template, err := template2.FromClientStack(stack)
	if err != nil {
		return err
	}
	template.Content = content

	files, err := template.RequiredFiles()
	if err != nil {
		return err
	}

	if err := readFiles(file, files, template, promptReplaceFile); err != nil {
		return err
	}

	if err := template.Validate(); err != nil {
		return fmt.Errorf("failed to parse template. If you are using go templating the template must execute with no values: %v", err)
	}

	mergeAnswers(template, answers)

	if err := populateAnswersFromEnv(template); err != nil {
		return err
	}

	if err := populateAnswersFromQuestions(template, prompt); err != nil {
		return err
	}

	newStack, err := template.ToClientStack()
	if err != nil {
		return err
	}

	_, err = wc.Stack.Update(stack, newStack)
	return err
}

func mergeAnswers(template *template2.Template, answers map[string]string) {
	for k, v := range answers {
		template.Answers[k] = v
	}
}

func populateAnswersFromEnv(template *template2.Template) error {
	keys, err := template.RequiredEnv()
	if err != nil {
		return err
	}

	for _, key := range keys {
		value := os.Getenv(key)
		if value != "" {
			template.Answers[key] = value
		}
	}

	return nil
}

func populateAnswersFromQuestions(template *template2.Template, forcePrompt bool) error {
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

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
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/up/questions"
	template2 "github.com/rancher/rio/pkg/template"
)

// readFileInPath reads a starting file source, a required file name and an optional content map, returns the content of required file
// file source can be http url, local disks or docker images
func readRequiredFiles(fileSource, requiredFile string, contents map[string]string) ([]byte, error) {
	// http url
	if strings.HasPrefix(fileSource, "http") {
		base, err := url.Parse(fileSource)
		if err != nil {
			return nil, err
		}
		ref, err := url.Parse(requiredFile)
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

	// docker registry
	if _, err := os.Stat(fileSource); err != nil {
		return []byte(contents[requiredFile]), nil
	}

	// local disk
	f, err := symlink.FollowSymlinkInScope(requiredFile, fileSource)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadFile(f)
}

func readFiles(workingDir string, contents map[string]string, requiredFiles []string, template *template2.Template, promptReplaceFile bool) error {
	for _, requireFile := range requiredFiles {
		existingContent, exists := template.AdditionalFiles[requireFile]
		content, err := readRequiredFiles(workingDir, requireFile, contents)
		if err != nil {
			if exists {
				continue
			}
			return fmt.Errorf("failed to read file %s: %v", requireFile, err)
		}

		if exists && bytes.Compare(existingContent, content) != 0 {
			yn := true
			if promptReplaceFile {
				yn, err = questions.PromptBool(fmt.Sprintf("The contents of %s have changed, do you want to update", requireFile), false)
				if err != nil {
					return errors.Wrap(err, "failed to ask question")
				}
			}

			if yn {
				template.AdditionalFiles[requireFile] = content
			}
		} else if !exists {
			template.AdditionalFiles[requireFile] = content
		}
	}

	return nil
}

func Run(ctx *clicontext.CLIContext, contents map[string]string, stackID string, promptReplaceFile, prompt bool, answers map[string]string, fileRef string) error {
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
	template.Content = []byte(contents[util.StackFileKey])

	requiredFiles, err := template.RequiredFiles()
	if err != nil {
		return err
	}

	wd, err := getWorkingDir(fileRef)
	if err != nil {
		return err
	}

	if err := readFiles(wd, contents, requiredFiles, template, promptReplaceFile); err != nil {
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

// getFileRef gets a file reference
func getWorkingDir(fileRef string) (string, error) {
	if _, err := os.Stat(fileRef); err != nil || strings.HasPrefix(fileRef, "http") {
		return fileRef, nil
	}
	return os.Getwd()
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

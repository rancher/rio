package apply

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/mapper/convert"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/up"
	"github.com/rancher/wrangler/pkg/yaml"
	"github.com/sirupsen/logrus"
)

type Apply struct {
	A_Answers string `desc:"Answer file in with key/value pairs in yaml or json"`
	Prompt    bool   `desc:"Re-ask all questions if answer is not found in environment variables"`
}

func (u *Apply) Run(ctx *clicontext.CLIContext) error {
	args := ctx.CLI.Args()
	if len(args) > 2 {
		return fmt.Errorf("either 1 or 2 arguements are required: [[STACK_NAME] FILE|-] or [DIRECTORY]")
	}

	switch len(args) {
	case 1:
		if stat, err := os.Stat(args[0]); err == nil && stat.IsDir() {
			return u.doUpAll(ctx, args[0])
		}
		return u.doUp(ctx, args[0], "")
	case 2:
		return u.doUp(ctx, args[1], args[0])
	default:
		panic("if you see this panic you have experienced something impossible")
	}
}

func (u *Apply) doUpAll(ctx *clicontext.CLIContext, dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), "-stack.yml") || strings.HasSuffix(f.Name(), "-stack.yaml") {
			if err := u.doUp(ctx, f.Name(), ""); err != nil {
				return err
			}
		}
	}

	return nil
}

func (u *Apply) doUp(ctx *clicontext.CLIContext, file, namespace string) error {
	content, err := util.ReadFile(file)
	if err != nil {
		return errors.Wrapf(err, "reading %s", file)
	}

	namespace, err = getNamespace(file, namespace)
	if err != nil {
		return err
	}

	answers, err := ReadAnswers(u.A_Answers)
	if err != nil {
		return fmt.Errorf("failed to parse answer file [%s]: %v", u.A_Answers, err)
	}

	logrus.Infof("Deploying rio-file to namespace [%s] from %s", namespace, file)
	if err := up.Run(ctx, content, namespace, answers); err != nil {
		return err
	}

	return nil
}

func ReadAnswers(answersFile string) (map[string]string, error) {
	content, err := util.ReadFile(answersFile)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{}
	if err := yaml.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	result := map[string]string{}
	for k, v := range data {
		result[k] = convert.ToString(v)
	}

	return result, nil
}

func getNamespace(file, namespace string) (string, error) {
	if namespace != "" {
		return namespace, nil
	}
	if strings.HasSuffix(file, "-stack.yml") || strings.HasSuffix(file, "-stack.yaml") {
		file = strings.TrimSuffix(file, "-stack.yml")
		file = strings.TrimSuffix(file, "-stack.yaml")
		return filepath.Base(file), nil
	}

	return "", fmt.Errorf("failed to determine stack name, please pass stack name as arguement")
}

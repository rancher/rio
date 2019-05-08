package apply

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/rancher/mapper/convert"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/up"
	"github.com/rancher/wrangler/pkg/yaml"
	"github.com/sirupsen/logrus"
)

type Apply struct {
	N_Namepsace string `desc:"Namespace to apply" default:"default"`
	F_File      string `desc:"the path to the rio file to apply"`
	A_Answers   string `desc:"Answer file in with key/value pairs in yaml or json"`
	Prompt      bool   `desc:"Re-ask all questions if answer is not found in environment variables"`
}

func (u *Apply) Run(ctx *clicontext.CLIContext) error {
	if u.F_File == "" {
		return fmt.Errorf("must specify filename")
	}
	return u.doUp(ctx, u.F_File, u.N_Namepsace)
}

func (u *Apply) doUp(ctx *clicontext.CLIContext, file, namespace string) error {
	content, err := util.ReadFile(file)
	if err != nil {
		return errors.Wrapf(err, "reading %s", file)
	}

	answers, err := ReadAnswers(u.A_Answers)
	if err != nil {
		return fmt.Errorf("failed to parse answer file [%s]: %v", u.A_Answers, err)
	}

	logrus.Infof("Deploying rio-file to namespace [%s] from %s", namespace, file)
	if err := up.Run(ctx, content, u.N_Namepsace, answers, u.Prompt); err != nil {
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

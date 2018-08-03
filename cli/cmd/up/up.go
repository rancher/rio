package up

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/norman/types/convert"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/up"
	"github.com/rancher/rio/cli/pkg/waiter"
	"github.com/rancher/rio/cli/server"
	"github.com/rancher/rio/pkg/yaml"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type Up struct {
	A_Answers string `desc:"Answer file in with key/value pairs in yaml or json"`
	Prompt    bool   `desc:"Re-ask all questions if answer is not found in environment variables"`
}

func (u *Up) Run(app *cli.Context) error {
	ctx, err := server.NewContext(app)
	if err != nil {
		return err
	}
	defer ctx.Close()

	args := app.Args()
	if len(args) > 2 {
		return fmt.Errorf("either 1 or 2 arguements are required: [[STACK_NAME] FILE|-] or [DIRECTORY]")
	}

	switch len(args) {
	case 1:
		if stat, err := os.Stat(args[0]); err == nil && stat.IsDir() {
			return u.doUpAll(ctx, args[0])
		} else {
			return u.doUp(ctx, args[0], "")
		}
	case 2:
		return u.doUp(ctx, args[1], args[0])
	default:
		panic("if you see this panic you have experienced something impossible")
	}
}

func (u *Up) doUpAll(ctx *server.Context, dir string) error {
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

func (u *Up) doUp(ctx *server.Context, file, stack string) error {
	content, err := util.ReadFile(file)
	if err != nil {
		return errors.Wrapf(err, "reading %s", file)
	}

	stack, err = getStackName(file, stack)
	if err != nil {
		return err
	}

	stackName := stack

	if len(stack) > 0 && !strings.HasSuffix(stack, "/") {
		stack += "/"
	}

	_, stackID, _, err := ctx.ResolveSpaceStackName(stack)
	if err != nil {
		return err
	}

	answers, err := readAnswers(u.A_Answers)
	if err != nil {
		return fmt.Errorf("failed to parse answer file [%s]: %v", u.A_Answers, err)
	}

	logrus.Infof("Deploying stack [%s] from %s", stackName, file)
	if err := up.Run(ctx, content, stackID, false, u.Prompt, answers); err != nil {
		return err
	}

	return waiter.WaitFor(ctx, stackID)
}

func readAnswers(answersFile string) (map[string]string, error) {
	content, err := util.ReadFile(answersFile)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	data, err := yaml.Parse(content)
	if err != nil {
		return nil, err
	}

	result := map[string]string{}
	for k, v := range data {
		result[k] = convert.ToString(v)
	}

	return result, nil
}

func getStackName(file, stack string) (string, error) {
	if stack != "" {
		return stack, nil
	}
	if strings.HasSuffix(file, "-stack.yml") || strings.HasSuffix(file, "-stack.yaml") {
		file = strings.TrimSuffix(file, "-stack.yml")
		file = strings.TrimSuffix(file, "-stack.yaml")
		return file, nil
	}

	return "", fmt.Errorf("failed to determine stack name, please pass stack name as arguement")
}

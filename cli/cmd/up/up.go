package up

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/rancher/rio/cli/cmd/apply"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/localbuilder"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/stack"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Up struct {
	Name        string `desc:"Set stack name, defaults to current directory name"`
	N_Namespace string `desc:"Set namespace" default:"default"`
	Answers     string `desc:"Set answer file"`
	F_File      string `desc:"Set rio file"`
	P_Parallel  bool   `desc:"Run builds in parallel"`
}

const (
	defaultRiofile       = "Riofile"
	defaultRiofileAnswer = "Riofile-answers"
)

func (u *Up) Run(c *clicontext.CLIContext) error {
	if u.Name == "" {
		u.Name = getCurrentDir()
	}

	if u.F_File == "" {
		if _, err := os.Stat(defaultRiofile); err == nil {
			u.F_File = defaultRiofile
		}
		if u.F_File == "" {
			return fmt.Errorf("can not found Riofile under current directory, must specify one. Example: rio up -f /path/to/Riofile.yaml")
		}
	}

	if u.Answers == "" {
		if _, err := os.Stat(defaultRiofileAnswer); err == nil {
			u.Answers = defaultRiofileAnswer
		}
	}
	answers, err := apply.ReadAnswers(u.Answers)
	if err != nil {
		return err
	}

	content, err := ioutil.ReadFile(u.F_File)
	if err != nil {
		return err
	}

	s := riov1.NewStack(u.N_Namespace, u.Name, riov1.Stack{
		Spec: riov1.StackSpec{
			Template: string(content),
			Answers:  answers,
		},
	})

	deployStack := stack.NewStack(content, answers)
	imageBuilds, err := deployStack.GetImageBuilds()
	if err != nil {
		return err
	}

	if len(imageBuilds) > 0 {
		localBuilder, err := localbuilder.NewLocalBuilder(c.Ctx, c.Apply, c.K8s)
		if err != nil {
			return err
		}

		images, err := localBuilder.Build(c.Ctx, imageBuilds, u.P_Parallel, u.N_Namespace)
		if err != nil {
			return err
		}
		s.Spec.Images = images
	}

	existing, err := c.Rio.Stacks(u.N_Namespace).Get(u.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return c.Create(s)
		}
	}
	existing.Spec = s.Spec
	return c.UpdateObject(existing)
}

func getCurrentDir() string {
	workingDir, _ := os.Getwd()
	dir := filepath.Base(workingDir)
	return dir
}

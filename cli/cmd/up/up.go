package up

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rancher/rio/cli/cmd/apply"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/localbuilder"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/stack"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Up struct {
	Name       string `desc:"Set stack name, defaults to current directory name"`
	Answers    string `desc:"Set answer file"`
	F_File     string `desc:"Set rio file"`
	P_Parallel bool   `desc:"Run builds in parallel"`
	Branch     string `desc:"Set branch when pointing stack to git repo" default:"master"`
	Revision   string `desc:"Set revision"`
}

const (
	defaultRiofile       = "Riofile"
	defaultRiofileAnswer = "Riofile-answers"
)

func (u *Up) Run(c *clicontext.CLIContext) error {
	if u.Name == "" {
		u.Name = getCurrentDir()
	}

	if len(c.CLI.Args()) > 0 {
		return u.setBuild(c)
	}

	content, answer, err := u.loadFile(c)
	if err != nil {
		return err
	}

	if err := u.updateStack(content, answer, c); err != nil {
		return err
	}

	existing, err := c.Rio.Stacks(c.GetSetNamespace()).Get(u.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	return u.up(content, answer, existing, c)
}

func (u *Up) setBuild(c *clicontext.CLIContext) error {
	s := riov1.NewStack(c.GetSetNamespace(), u.Name, riov1.Stack{
		Spec: riov1.StackSpec{
			Build: &riov1.StackBuild{
				Repo:     c.CLI.Args()[0],
				Branch:   u.Branch,
				Revision: u.Revision,
			},
		},
	})

	existing, err := c.Rio.Stacks(c.GetSetNamespace()).Get(u.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return c.Create(s)
		}
	}
	if existing.Spec.Build == nil {
		existing.Spec.Build = &riov1.StackBuild{}
	}
	existing.Spec.Build.Repo = c.CLI.Args()[0]
	existing.Spec.Build.Branch = u.Branch
	existing.Spec.Build.Revision = u.Revision
	return c.UpdateObject(existing)
}

func (u *Up) loadFile(c *clicontext.CLIContext) (string, map[string]string, error) {
	if u.F_File == "" {
		if _, err := os.Stat(defaultRiofile); err == nil {
			u.F_File = defaultRiofile
		}
		if u.F_File == "" {
			return "", nil, fmt.Errorf("can not found Riofile under current directory, must specify one. Example: rio up -f /path/to/Riofile.yaml")
		}
	}

	if u.Answers == "" {
		if _, err := os.Stat(defaultRiofileAnswer); err == nil {
			u.Answers = defaultRiofileAnswer
		}
	}
	answers, err := apply.ReadAnswers(u.Answers)
	if err != nil {
		return "", nil, err
	}

	content, err := readFile(u.F_File)
	if err != nil {
		return "", nil, err
	}
	return string(content), answers, nil
}

func (u *Up) up(content string, answers map[string]string, s *riov1.Stack, c *clicontext.CLIContext) error {
	deployStack := stack.NewStack([]byte(content), answers)
	imageBuilds, err := deployStack.GetImageBuilds()
	if err != nil {
		return err
	}

	if len(imageBuilds) > 0 {
		localBuilder, err := localbuilder.NewLocalBuilder(c.Ctx, c.SystemNamespace, c.Apply, c.K8s)
		if err != nil {
			return err
		}

		images, err := localBuilder.Build(c.Ctx, imageBuilds, u.P_Parallel, c.GetSetNamespace())
		if err != nil {
			return err
		}
		for k, i := range images {
			localRegsitry := constants.RegistryService
			if strings.HasPrefix(i, localRegsitry) {
				images[k] = strings.Replace(i, localRegsitry, "localhost:5442", -1)
			}
		}
		if err := deployStack.SetServiceImages(images); err != nil {
			return err
		}
	}
	objs, err := deployStack.GetObjects()
	if err != nil {
		return err
	}
	return c.Apply.WithDefaultNamespace(c.GetSetNamespace()).WithOwner(s).WithSetOwnerReference(true, true).WithDynamicLookup().ApplyObjects(objs...)
}

func (u *Up) updateStack(content string, answers map[string]string, c *clicontext.CLIContext) error {
	s := riov1.NewStack(c.GetSetNamespace(), u.Name, riov1.Stack{
		Spec: riov1.StackSpec{
			Template: content,
			Answers:  answers,
		},
	})

	existing, err := c.Rio.Stacks(c.GetSetNamespace()).Get(u.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return c.Create(s)
		}
		return err
	}
	existing.Spec.Template = s.Spec.Template
	existing.Spec.Answers = s.Spec.Answers
	return c.UpdateObject(existing)
}

func readFile(file string) ([]byte, error) {
	if file == "-" {
		return ioutil.ReadAll(os.Stdin)
	}
	if strings.HasPrefix(file, "http") {
		resp, err := http.Get(file)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)
	}
	return ioutil.ReadFile(file)
}

func getCurrentDir() string {
	workingDir, _ := os.Getwd()
	dir := filepath.Base(workingDir)
	return dir
}

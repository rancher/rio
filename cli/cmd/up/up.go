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
	namespace := c.GetSetNamespace()
	if namespace == "" {
		namespace = c.GetDefaultNamespace()
	}

	if len(c.CLI.Args()) > 0 {
		s := riov1.NewStack(namespace, u.Name, riov1.Stack{
			Spec: riov1.StackSpec{
				Build: &riov1.StackBuild{
					Repo:     c.CLI.Args()[0],
					Branch:   u.Branch,
					Revision: u.Revision,
				},
			},
		})

		existing, err := c.Rio.Stacks(namespace).Get(u.Name, metav1.GetOptions{})
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

	content, err := readFile(u.F_File)
	if err != nil {
		return err
	}

	s := riov1.NewStack(namespace, u.Name, riov1.Stack{
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
		pushLocal := os.Getenv("PUSH_LOCAL") == "TRUE"
		localBuilder, err := localbuilder.NewLocalBuilder(c.Ctx, c.BuildkitPodName, c.SocatPodName, c.SystemNamespace, pushLocal, c.Apply, c.K8s)
		if err != nil {
			return err
		}

		images, err := localBuilder.Build(c.Ctx, imageBuilds, u.P_Parallel, namespace)
		if err != nil {
			return err
		}
		for k, i := range images {
			localRegsitry := fmt.Sprintf("registry.%s", c.SystemNamespace)
			if strings.HasPrefix(i, localRegsitry) {
				images[k] = strings.Replace(i, localRegsitry, "localhost:5442", -1)
			}
		}
		s.Spec.Images = images
	}

	existing, err := c.Rio.Stacks(namespace).Get(u.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return c.Create(s)
		}
		return err
	}
	existing.Spec.Template = s.Spec.Template
	existing.Spec.Answers = s.Spec.Answers
	existing.Spec.Images = s.Spec.Images
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

package clientcfg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/docker/docker/pkg/reexec"
	"github.com/rancher/norman/objectclient/dynamic"
	"github.com/rancher/norman/restwatch"
	"github.com/rancher/rio/cli/pkg/up/questions"
	"github.com/rancher/rio/pkg/clientaccess"
	"github.com/rancher/rio/pkg/project"
	"github.com/rancher/rio/pkg/settings"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	corev1 "github.com/rancher/types/apis/core/v1"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type Cluster struct {
	clientaccess.Info

	ID                 string  `json:"id,omitempty"`
	Checksum           string  `json:"-"`
	Name               string  `json:"name,omitempty"`
	DefaultStackName   string  `json:"defaultStackName,omitempty"`
	DefaultProjectName string  `json:"defaultProjectName,omitempty"`
	Default            bool    `json:"default,omitempty"`
	Config             *Config `json:"-"`

	File    string `json:"-"`
	domain  string
	project *Project
}

func (c *Cluster) Project() (*Project, error) {
	if c.Config.ProjectName != "" {
		projects, err := c.projects(true)
		if err != nil {
			return nil, err
		}

		for _, w := range projects {
			if w.Project.Name == c.Config.ProjectName {
				return &w, nil
			}
		}

		return nil, fmt.Errorf("failed to find project %s", c.Config.ProjectName)
	}

	projects, err := c.Projects()
	if err != nil {
		return nil, err
	}

	for _, w := range projects {
		if w.Project.Name == c.DefaultProjectName {
			return &w, nil
		}
	}

	if len(projects) == 0 {
		return c.CreateProject(c.DefaultProjectName)
	}

	if len(projects) == 1 {
		return &projects[0], nil
	}

	msg := "Choose a project (run 'rio set-context' to set default):\n"
	var options []string

	for i, w := range projects {
		msg := fmt.Sprintf("[%d] %s\n", i+1, w.Project.Name)
		options = append(options, msg)
	}

	choice, err := questions.PromptOptions(msg, -1, options...)
	if err != nil {
		return nil, err
	}

	c.project = &projects[choice]
	return c.project, nil
}

func (c *Cluster) CreateProject(name string) (*Project, error) {
	client, err := c.KubeClient()
	if err != nil {
		return nil, err
	}
	projectNs := corev1.NewNamespace("", name, v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				project.ProjectLabel: "true",
			},
		},
	})
	ns, err := client.Core.Namespaces("").Create(projectNs)
	if err != nil {
		return nil, err
	}
	return &Project{
		Project: ns,
		Cluster: c,
	}, nil
}

type KubeClient struct {
	Core    corev1.Interface
	Rio     riov1.Interface
	Project projectv1.Interface
}

func (c *Cluster) RestClient() (rest.Interface, error) {
	config := c.Info.RestConfig()
	if config.NegotiatedSerializer == nil {
		config.NegotiatedSerializer = dynamic.NegotiatedSerializer
	}

	return restwatch.UnversionedRESTClientFor(config)
}

func (c *Cluster) KubeClient() (*KubeClient, error) {
	config := c.Info.RestConfig()

	coreClient, err := corev1.NewForConfig(*config)
	if err != nil {
		return nil, err
	}

	rioClient, err := riov1.NewForConfig(*config)
	if err != nil {
		return nil, err
	}

	projectClient, err := projectv1.NewForConfig(*config)
	if err != nil {
		return nil, err
	}

	return &KubeClient{
		Core:    coreClient,
		Rio:     rioClient,
		Project: projectClient,
	}, nil
}

func (c *Cluster) Projects() ([]Project, error) {
	return c.projects(false)
}

func (c *Cluster) projects(all bool) ([]Project, error) {
	client, err := c.KubeClient()
	if err != nil {
		return nil, err
	}
	projects, err := client.Core.Namespaces("").List(metav1.ListOptions{
		LabelSelector: "rio.cattle.io/project=true",
	})
	if err != nil {
		return nil, err
	}

	var result []Project
	for _, p := range projects.Items {
		if !all && p.Name == settings.RioSystemNamespace {
			continue
		}
		project := c.projectFromSpace(p)
		if p.Name == c.DefaultProjectName {
			project.Default = true
		}
		result = append(result, *project)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Project.Name < result[j].Project.Name
	})

	return result, nil
}

func (c *Cluster) projectFromSpace(project v1.Namespace) *Project {
	return &Project{
		Project: &project,
		Cluster: c,
	}
}

func (c *Cluster) KubectlCmd(namespace, command string, args ...string) (*exec.Cmd, error) {
	var err error

	kc, err := c.kubeConfig()
	if err != nil {
		return nil, err
	}

	execArgs := []string{"kubectl", fmt.Sprintf("--kubeconfig=%s", kc)}
	if logrus.GetLevel() >= logrus.DebugLevel {
		execArgs = append(execArgs, "-v=9")
	}
	if namespace != "" {
		execArgs = append(execArgs, "-n", namespace)
	}
	if command != "" {
		execArgs = append(execArgs, command)
	}
	execArgs = append(execArgs, args...)

	logrus.Debugf("%v, KUBECONFIG=%s", execArgs, kc)
	cmd := reexec.Command(execArgs...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", kc))
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd, nil
}

func (c *Cluster) Kubectl(namespace, command string, args ...string) error {
	cmd, err := c.KubectlCmd(namespace, command, args...)
	if err != nil {
		return err
	}
	return cmd.Run()
}

func (c *Cluster) kubeConfig() (string, error) {
	kcc := c.Config.KubeconfigCache()
	kc := filepath.Join(kcc, c.Checksum)
	if _, err := os.Stat(kc); err == nil {
		return kc, nil
	}

	return kc, c.Info.WriteKubeConfig(kc)
}

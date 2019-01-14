package clientcfg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/docker/docker/pkg/reexec"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/rio/cli/pkg/up/questions"
	"github.com/rancher/rio/pkg/clientaccess"
	"github.com/rancher/rio/pkg/settings"
	projectclient "github.com/rancher/rio/types/client/project/v1"
	"github.com/sirupsen/logrus"
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

	File          string `json:"-"`
	domain        string
	project       *Project
	clientInfo    *clusterClientInfo
	projectClient *projectclient.Client
}

func (c *Cluster) Client() (*projectclient.Client, error) {
	if c.projectClient != nil {
		return c.projectClient, nil
	}

	info, err := c.getClientInfo()
	if err != nil {
		return nil, err
	}

	sc, err := info.clusterClient()
	if err != nil {
		return nil, err
	}

	c.projectClient = sc
	return c.projectClient, nil
}

func (c *Cluster) getClientInfo() (*clusterClientInfo, error) {
	if c.clientInfo != nil {
		return c.clientInfo, nil
	}

	cci, err := newClusterClientInfo(&c.Info)
	if err != nil {
		return nil, err
	}

	c.clientInfo = cci
	return cci, nil
}

func (c *Cluster) Project() (*Project, error) {
	if c.Config.ProjectName != "" {
		projects, err := c.projects(true)
		if err != nil {
			return nil, err
		}

		for _, w := range projects {
			if w.ID == c.Config.ProjectName || w.Name == c.Config.ProjectName {
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
		if w.ID == c.DefaultProjectName || w.Name == c.DefaultProjectName {
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
		msg := fmt.Sprintf("[%d] %s\n", i+1, w.Name)
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
	sc, err := c.Client()
	if err != nil {
		return nil, err
	}

	space, err := sc.Project.Create(&projectclient.Project{
		Name: name,
	})
	if err != nil {
		return nil, err
	}

	return c.projectFromSpace(*space), nil
}

func (c *Cluster) Projects() ([]Project, error) {
	return c.projects(false)
}

func (c *Cluster) projects(all bool) ([]Project, error) {
	sc, err := c.Client()
	if err != nil {
		return nil, err
	}

	projects, err := sc.Project.List(util.DefaultListOpts())
	if err != nil {
		return nil, err
	}

	var result []Project
	for _, p := range projects.Data {
		if !all && p.ID == settings.RioSystemNamespace {
			continue
		}
		project := c.projectFromSpace(p)
		if p.Name == c.DefaultProjectName {
			project.Default = true
		}
		result = append(result, *project)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

func (c *Cluster) projectFromSpace(project projectclient.Project) *Project {
	return &Project{
		Project: project,
		Cluster: c,
	}
}

func (c *Cluster) Domain() (string, error) {
	if c.domain != "" {
		return c.domain, nil
	}

	ci, err := c.getClientInfo()
	if err != nil {
		return "", err
	}

	domain, err := ci.Domain()
	if err != nil {
		return "", err
	}

	c.domain = domain
	return c.domain, nil
}

func (c *Cluster) KubectlCmd(namespace, command string, args ...string) (*exec.Cmd, error) {
	var err error

	kc := os.Getenv(fmt.Sprintf("KUBECONFIG_%s_DEV", c.Name))
	if kc == "" {
		kc, err = c.kubeConfig()
		if err != nil {
			return nil, err
		}
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

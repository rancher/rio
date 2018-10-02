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
	spaceclient "github.com/rancher/rio/types/client/space/v1beta1"
	"github.com/sirupsen/logrus"
)

type Cluster struct {
	clientaccess.Info

	ID                   string  `json:"id,omitempty"`
	Checksum             string  `json:"-"`
	Name                 string  `json:"name,omitempty"`
	DefaultStackName     string  `json:"defaultStackName,omitempty"`
	DefaultWorkspaceName string  `json:"defaultWorkspaceName,omitempty"`
	Default              bool    `json:"-"`
	Config               *Config `json:"-"`

	file        string
	domain      string
	workspace   *Workspace
	clientInfo  *clusterClientInfo
	spaceClient *spaceclient.Client
}

func (c *Cluster) Client() (*spaceclient.Client, error) {
	if c.spaceClient != nil {
		return c.spaceClient, nil
	}

	info, err := c.getClientInfo()
	if err != nil {
		return nil, err
	}

	sc, err := info.spaceClient()
	if err != nil {
		return nil, err
	}

	c.spaceClient = sc
	return c.spaceClient, nil
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

func (c *Cluster) Workspace() (*Workspace, error) {
	if c.Config.WorkspaceName != "" {
		workspaces, err := c.workspaces(true)
		if err != nil {
			return nil, err
		}

		for _, w := range workspaces {
			if w.ID == c.Config.WorkspaceName || w.Name == c.Config.WorkspaceName {
				return &w, nil
			}
		}

		return nil, fmt.Errorf("failed to find workspace %s", c.Config.WorkspaceName)
	}

	workspaces, err := c.Workspaces()
	if err != nil {
		return nil, err
	}

	for _, w := range workspaces {
		if w.ID == c.DefaultWorkspaceName || w.Name == c.DefaultWorkspaceName {
			return &w, nil
		}
	}

	if len(workspaces) == 0 {
		return c.CreateWorkspace(c.DefaultWorkspaceName)
	}

	if len(workspaces) == 1 {
		return &workspaces[0], nil
	}

	msg := "Choose a workspace (run 'rio set-context' to set default):\n"
	var options []string

	for i, w := range workspaces {
		msg := fmt.Sprintf("[%d] %s\n", i+1, w.Name)
		options = append(options, msg)
	}

	choice, err := questions.PromptOptions(msg, -1, options...)
	if err != nil {
		return nil, err
	}

	c.workspace = &workspaces[choice]
	return c.workspace, nil
}

func (c *Cluster) CreateWorkspace(name string) (*Workspace, error) {
	sc, err := c.Client()
	if err != nil {
		return nil, err
	}

	space, err := sc.Space.Create(&spaceclient.Space{
		Name: name,
	})
	if err != nil {
		return nil, err
	}

	return c.workspaceFromSpace(*space), nil
}

func (c *Cluster) Workspaces() ([]Workspace, error) {
	return c.workspaces(false)
}

func (c *Cluster) workspaces(all bool) ([]Workspace, error) {
	sc, err := c.Client()
	if err != nil {
		return nil, err
	}

	workspaces, err := sc.Space.List(util.DefaultListOpts())
	if err != nil {
		return nil, err
	}

	var result []Workspace
	for _, w := range workspaces.Data {
		if !all && w.ID == settings.RioSystemNamespace {
			continue
		}
		workspace := c.workspaceFromSpace(w)
		if w.Name == c.DefaultWorkspaceName {
			workspace.Default = true
		}
		result = append(result, *workspace)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

func (c *Cluster) workspaceFromSpace(space spaceclient.Space) *Workspace {
	return &Workspace{
		Space:   space,
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

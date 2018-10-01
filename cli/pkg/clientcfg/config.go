package clientcfg

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/pkg/resolvehome"
	"github.com/rancher/rio/cli/pkg/up/questions"
	"github.com/rancher/rio/pkg/clientaccess"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	spaceclient "github.com/rancher/rio/types/client/space/v1beta1"
	"github.com/sirupsen/logrus"
)

var ErrNoConfig = errors.New("no config found")

type Config struct {
	Home          string
	ClusterName   string
	WorkspaceName string
	Debug         bool
	Wait          bool
	WaitTimeout   int
	WaitState     string
	ServerURL     string
	Token         string

	cluster *Cluster
}

func (c *Config) ClusterDir() string {
	return filepath.Join(c.ClientDir(), "clusters")
}

func (c *Config) KubeconfigCache() string {
	return filepath.Join(c.ClientDir(), "kubeconfig-cache")
}

func (c *Config) ClientDir() string {
	return filepath.Join(c.Home, "client")
}

func (c *Config) DefaultClusterName() (string, error) {
	return defaultNameFromFile(filepath.Join(c.ClientDir(), "default-cluster"))
}

func (c *Config) Validate() error {
	if c.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	home, err := resolvehome.Resolve(c.Home)
	if err != nil {
		return err
	}
	c.Home = home
	return nil
}

func (c *Config) Workspace() (*Workspace, error) {
	cluster, err := c.Cluster()
	if err != nil {
		return nil, err
	}
	return cluster.Workspace()
}

func (c *Config) WorkspaceClient() (*client.Client, error) {
	w, err := c.Workspace()
	if err != nil {
		return nil, err
	}
	return w.Client()
}

func (c *Config) ClusterClient() (*spaceclient.Client, error) {
	cluster, err := c.Cluster()
	if err != nil {
		return nil, err
	}
	return cluster.Client()
}

func (c *Config) Cluster() (*Cluster, error) {
	if c.cluster != nil {
		return c.cluster, nil
	}

	clusters, err := c.Clusters()
	if err != nil {
		return nil, err
	}

	if c.ClusterName != "" {
		for i, cluster := range clusters {
			if cluster.ID == c.ClusterName || cluster.Name == c.ClusterName {
				return &clusters[i], nil
			}
		}
	}

	var defaultCluster *Cluster
	for i, cluster := range clusters {
		if cluster.Default {
			defaultCluster = &clusters[i]
			break
		}
	}

	if defaultCluster != nil {
		return defaultCluster, nil
	}

	if len(clusters) == 0 {
		adminCluster := c.getAndSaveAdminCluster()
		if adminCluster != nil {
			return adminCluster, nil
		}
		return nil, ErrNoConfig
	}

	if len(clusters) == 1 {
		return &clusters[0], nil
	}

	msg := "Choose a cluster (run 'rio set-context' to set default):\n"
	var options []string

	for i, c := range clusters {
		name := c.Name
		if c.ID != "" && c.Name != c.ID {
			name = fmt.Sprintf("%s(%s)", c.Name, c.ID)
		}
		msg := fmt.Sprintf("[%d] %s %s\n", i+1, name, c.URL.String())
		options = append(options, msg)
	}

	choice, err := questions.PromptOptions(msg, -1, options...)
	if err != nil {
		return nil, err
	}

	c.cluster = &clusters[choice]
	return c.cluster, nil
}

func (c *Config) Clusters() ([]Cluster, error) {
	clusters, err := listClusters(c.ClusterDir())
	if err != nil {
		return nil, err
	}

	defaultName, err := c.DefaultClusterName()
	if err != nil {
		return nil, err
	}

	for i := range clusters {
		if clusters[i].DefaultWorkspaceName == "" {
			clusters[i].DefaultWorkspaceName = "default"
		}
		clusters[i].Default = clusters[i].Name == defaultName
		clusters[i].Config = c
	}

	return clusters, nil
}

func (c *Config) SaveCluster(cluster *Cluster) error {
	_, err := cluster.Workspaces()
	if err != nil {
		return errors.Wrapf(err, "can not save cluster")
	}

	name := cluster.ID
	if name == "" {
		name = cluster.Name
	}
	if name == "" {
		return fmt.Errorf("can not save cluster with no name or ID")
	}

	clusterDir := c.ClientDir()
	file := filepath.Join(clusterDir, fmt.Sprintf("%s.json", name))

	out, err := os.Create(file)
	if err != nil {
		return errors.Wrapf(err, "can not save cluster")
	}

	err = json.NewEncoder(out).Encode(cluster)
	if err != nil {
		return errors.Wrapf(err, "can not save cluster")
	}

	return nil
}

func (c *Config) getAndSaveAdminCluster() *Cluster {
	bytes, err := ioutil.ReadFile("/var/lib/rancher/rio/server/port")
	if err != nil {
		return nil
	}
	token, err := ioutil.ReadFile("/var/lib/rancher/rio/server/client-token")
	if err != nil {
		return nil
	}

	url := fmt.Sprintf("https://localhost:%s", bytes)

	info, err := clientaccess.ParseAndValidateToken(url, strings.TrimSpace(string(token)))
	if err != nil {
		return nil
	}

	info.Token = strings.TrimSpace(string(token))
	cluster := &Cluster{
		Info:                 *info,
		DefaultStackName:     "default",
		DefaultWorkspaceName: "default",
		ID:                   "local",
		Name:                 "local",
		Checksum:             "local",
		Config:               c,
	}

	if err := c.SaveCluster(cluster); err != nil {
		return nil
	}

	return cluster
}

func listClusters(dir string) ([]Cluster, error) {
	files, err := ioutil.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	var result []Cluster
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		fName := filepath.Join(dir, file.Name())
		data, err := ioutil.ReadFile(fName)
		if err != nil {
			logrus.Errorf("Failed to open cluster config %s: %v", fName, err)
			continue
		}

		digest := sha256.Sum256(data)
		cluster := Cluster{}
		if err := json.Unmarshal(data, &cluster); err != nil {
			logrus.Errorf("Failed to parse cluster config %s: %v", fName, err)
			continue
		}

		cluster.file = fName
		cluster.Checksum = hex.EncodeToString(digest[:])

		if cluster.Name == "" {
			cluster.Name = strings.TrimSuffix(file.Name(), ".json")
		}

		if cluster.DefaultStackName == "" {
			cluster.DefaultStackName = "default"
		}

		result = append(result, cluster)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

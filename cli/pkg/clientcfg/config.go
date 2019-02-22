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
	"github.com/rancher/norman/clientbase"
	"github.com/rancher/norman/pkg/resolvehome"
	"github.com/rancher/rio/cli/pkg/up/questions"
	"github.com/rancher/rio/pkg/clientaccess"
	"github.com/rancher/rio/pkg/settings"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

const (
	defaultProjectName = "rio"
)

var ErrNoConfig = errors.New("no config found")

type Config struct {
	Home        string
	ClusterName string
	ProjectName string
	Debug       bool
	Wait        bool
	WaitTimeout int
	WaitState   string
	ServerURL   string
	Token       string

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
		clientbase.Debug = true
	}
	home, err := resolvehome.Resolve(c.Home)
	if err != nil {
		return err
	}
	c.Home = home
	return nil
}

func (c *Config) KubeClient() (*KubeClient, error) {
	cluster, err := c.Cluster()
	if err != nil {
		return nil, err
	}
	return cluster.KubeClient()
}

func (c *Config) Project() (*Project, error) {
	cluster, err := c.Cluster()
	if err != nil {
		return nil, err
	}
	return cluster.Project()
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
		return nil, fmt.Errorf("failed to find cluster %s", c.ClusterName)
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
		msg := fmt.Sprintf("[%d] %s %s\n", i+1, name, c.URL)
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
	clusters, err := ListClusters(c.ClusterDir())
	if err != nil {
		return nil, err
	}

	defaultName, err := c.DefaultClusterName()
	if err != nil {
		return nil, err
	}

	for i := range clusters {
		if clusters[i].DefaultProjectName == "" {
			clusters[i].DefaultProjectName = defaultProjectName
		}
		if !clusters[i].Default {
			clusters[i].Default = clusters[i].Name == defaultName
		}
		clusters[i].Config = c
	}

	return clusters, nil
}

func (c *Config) SaveCluster(cluster *Cluster, config *rest.Config) error {
	name := cluster.ID

	if config != nil {
		cluster.Info = clientaccess.Info{
			URL:      config.Host,
			CACerts:  config.CAData,
			Username: config.Username,
			Password: config.Password,
			Token:    config.BearerToken,
		}
	}

	clusterDir := c.ClusterDir()
	if err := os.MkdirAll(clusterDir, 0700); err != nil {
		return err
	}

	file := cluster.File
	if file == "" {
		file = filepath.Join(clusterDir, fmt.Sprintf("%s.json", name))
	}

	out, err := os.Create(file)
	if err != nil && !os.IsExist(err) {
		return errors.Wrapf(err, "can not save cluster")
	} else if os.IsExist(err) {
		out, err = os.Open(file)
		if err != nil {
			return errors.Wrapf(err, "can not update cluster")
		}
	}

	err = json.NewEncoder(out).Encode(cluster)
	if err != nil {
		return errors.Wrapf(err, "can not save cluster")
	}

	return nil
}

func ListClusters(dir string) ([]Cluster, error) {
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

		cluster.File = fName
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

func (c *Cluster) Domain() (string, error) {
	client, err := c.KubeClient()
	if err != nil {
		return "", err
	}

	if settings.ClusterDomain.Get() == "" {
		return "", nil
	}
	domain, err := client.Project.Settings("").Get(settings.ClusterDomain.Get(), metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return domain.Value, nil
}

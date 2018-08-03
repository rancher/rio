package server

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/rancher/norman/types"
	"github.com/rancher/rio/cli/pkg/kv"
	"github.com/rancher/rio/pkg/clientaccess"
	"github.com/rancher/rio/types/client/rio/v1beta1"
	spaceclient "github.com/rancher/rio/types/client/space/v1beta1"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var ErrNoConfig = errors.New("no config found")

type Context struct {
	CLIContext       *cli.Context
	Domain           string
	Client           *client.Client
	DefaultStackName string
	SpaceID          string
	builder          *ContextBuilder
	tempFile         string
}

func (c *Context) Close() error {
	if c.tempFile != "" {
		return os.Remove(c.tempFile)
	}
	return nil
}

func getAdminConfig() string {
	bytes, err := ioutil.ReadFile("/var/lib/rancher/rio/server/port")
	if err != nil {
		return ""
	}
	token, err := ioutil.ReadFile("/var/lib/rancher/rio/server/client-token")
	if err != nil {
		return ""
	}

	url := fmt.Sprintf("https://localhost:%s", bytes)
	tmpDir := "/var/lib/rancher/rio/server/tmp"
	os.MkdirAll(tmpDir, 0700)
	tempFile, err := clientaccess.AccessInfoToTempKubeConfig(tmpDir, url, strings.TrimSpace(string(token)))
	if err == nil {
		return tempFile
	}

	return ""
}

func getK8sConfig(app *cli.Context) string {
	kubeConfig := app.GlobalString("kubeconfig")

	if kubeConfig != "" {
		return kubeConfig
	}

	k8sConf, err := K8sConfPath()
	if err != nil {
		logrus.Errorf("Failed to determine kubeconfig path %s: %v", k8sConf, err)
		return ""
	}

	if _, err := os.Stat(k8sConf); err == nil {
		return k8sConf
	}

	return ""
}

func getRioConf(app *cli.Context) (string, bool, error) {
	ch, rio, _, err := Paths()
	if err != nil {
		return "", false, fmt.Errorf("failed to paths for config: %v", err)
	}

	serverURL := app.GlobalString("server")
	token := app.GlobalString("token")

	if serverURL == "" || token == "" {
		if _, err := os.Stat(rio); err == nil {
			return rio, false, nil
		}

		return "", false, nil
	}

	tempFile, err := clientaccess.AccessInfoToTempKubeConfig(ch, serverURL, token)
	return tempFile, true, err
}

func NewContext(app *cli.Context) (*Context, error) {
	k8s := false
	conf, temp, err := getRioConf(app)
	if err != nil {
		return nil, err
	}

	if conf == "" {
		conf = getK8sConfig(app)
		k8s = true
	}

	if conf == "" {
		conf = getAdminConfig()
		temp = conf != ""
		k8s = false
	}

	if conf == "" {
		return nil, ErrNoConfig
	}

	builder, err := NewContextBuilder(conf, k8s)
	if err != nil {
		return nil, err
	}

	domain, err := builder.Domain()
	if err != nil {
		return nil, errors.Wrap(err, "getting cluster domain")
	}

	workspace := app.GlobalString("workspace")
	client, err := builder.Client(workspace)
	if err != nil {
		return nil, errors.Wrap(err, "building client")
	}

	ctx := &Context{
		CLIContext:       app,
		Domain:           domain,
		Client:           client,
		DefaultStackName: "default",
		SpaceID:          workspace,
		builder:          builder,
	}

	if temp {
		ctx.tempFile = conf
	}

	dev := os.Getenv("KUBECONFIG_DEV")
	if !k8s && dev != "" {
		os.Setenv("KUBECONFIG", dev)
	} else {
		os.Setenv("KUBECONFIG", conf)
	}
	return ctx, nil
}

func (c *Context) SpaceClient() (*spaceclient.Client, error) {
	return c.builder.SpaceClient()
}

func (c *Context) ResolveSpaceStackName(in string) (string, string, string, error) {
	stackName, name := kv.Split(in, "/")
	if stackName != "" && name == "" {
		if !strings.HasSuffix(in, "/") {
			name = stackName
			stackName = ""
		}
	}

	if stackName == "" {
		stackName = c.DefaultStackName
	}

	stacks, err := c.Client.Stack.List(&types.ListOpts{
		Filters: map[string]interface{}{
			"name": stackName,
		},
	})
	if err != nil {
		return "", "", "", errors.Wrapf(err, "failed to determine stack")
	}

	var s *client.Stack
	if len(stacks.Data) == 0 {
		s, err = c.Client.Stack.Create(&client.Stack{
			Name:    stackName,
			SpaceID: c.SpaceID,
		})
		if err != nil {
			return "", "", "", errors.Wrapf(err, "failed to create stack %s", stackName)
		}
	} else {
		s = &stacks.Data[0]
	}

	return s.SpaceID, s.ID, name, nil
}

func SpaceClient(conf string, k8s bool) (*spaceclient.Client, error) {
	builder, err := NewContextBuilder(conf, k8s)
	if err != nil {
		return nil, err
	}
	return builder.SpaceClient()
}

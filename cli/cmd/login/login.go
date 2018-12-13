package login

import (
	"io/ioutil"
	"strings"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/clientcfg"
	"github.com/rancher/rio/cli/pkg/up/questions"
	"github.com/rancher/rio/pkg/clientaccess"
	"github.com/rancher/rio/pkg/name"
	"github.com/sirupsen/logrus"
)

type Login struct {
	S_Server string `desc:"Server to log into"`
	T_Token  string `desc:"Authentication token"`
}

func (l *Login) Run(ctx *clicontext.CLIContext) (ex error) {
	defer func() {
		if ex == nil {
			logrus.Infof("Log in successful")
		}
	}()

	var err error

	if l.S_Server == "" {
		l.S_Server, err = questions.Prompt("Rio server URL: ", "")
		if err != nil {
			return err
		}
	}

	if l.T_Token == "" {
		l.T_Token, err = questions.Prompt("Authentication token: ", "")
		if err != nil {
			return err
		}
	}

	bytes, err := ioutil.ReadFile(l.T_Token)
	if err == nil && len(bytes) > 0 {
		l.T_Token = strings.TrimSpace(string(bytes))
	}

	cluster, err := Validate(l.S_Server, l.T_Token)
	if err != nil {
		return err
	}

	cluster.ID = name.Hex(cluster.URL, 5)
	cluster.Name = cluster.ID
	return ctx.Config.SaveCluster(cluster, true)
}

func Validate(serverURL, token string) (*clientcfg.Cluster, error) {
	info, err := clientaccess.ParseAndValidateToken(serverURL, token)
	if err != nil {
		return nil, err
	}

	cluster := &clientcfg.Cluster{
		Info: *info,
	}

	if _, err := cluster.Client(); err != nil {
		return nil, err
	}

	return cluster, nil
}

package agent

import "github.com/urfave/cli"

type Agent struct {
	T_Token   string `desc:"Token to use for authentication" env:"RIO_TOKEN"`
	S_Server  string `desc:"Server to connect to" env:"RIO_URL"`
	D_DataDir string `desc:"Folder to hold state" default:"/var/lib/rancher/rio"`
	L_Log     string `desc:"log to file"`
}

func (a *Agent) Customize(command *cli.Command) {
	command.Category = "CLUSTER RUNTIME"
}

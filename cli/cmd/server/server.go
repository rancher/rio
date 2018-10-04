package server

import "github.com/urfave/cli"

type Server struct {
	P_HttpsListenPort  int    `desc:"HTTPS listen port" default:"7443"`
	L_HttpListenPort   int    `desc:"HTTP listen port" default:"7080"`
	D_DataDir          string `desc:"Folder to hold state default /var/lib/rancher/rio or ${HOME}/.rancher/rio if not root"`
	DisableControllers bool   `desc:"Don't run controllers (only useful for rio development)"`
	DisableAgent       bool   `desc:"Do not run a local agent and register this server"`
	ProfilePort        int    `desc:"Profiling port, 0 disables profiling" default:"6060"`
	Log                string `desc:"Log to file"`
}

func (s *Server) Customize(command *cli.Command) {
	command.Category = "CLUSTER RUNTIME"
}

package logs

import (
	"github.com/rancher/rio/cli/pkg/clicontext"
)

type Logs struct {
	F_Follow     bool   `desc:"Follow log output"`
	S_Since      string `desc:"Logs since a certain time, either duration (5s, 2m, 3h) or RFC3339"`
	P_Previous   bool   `desc:"Print the logs for the previous instance of the container in a pod if it exists"`
	C_Container  string `desc:"Print the logs of a specific container"`
	R_Revision   string `desc:"Print the logs of a specific revision"`
	N_Tail       int    `desc:"Number of recent lines of logs to print, -1 for all" default:"200"`
	A_All        bool   `desc:"Include hidden or systems logs when logging"`
	T_Timestamps bool   `desc:"Print the logs with timestamp"`

	Pod string `desc:"Include hidden or systems logs when logging"`
}

func (l *Logs) Run(ctx *clicontext.CLIContext) error {
}

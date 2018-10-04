// +build k3s

package agent

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/resolvehome"
	"github.com/rancher/rio/pkg/clientaccess"
	"github.com/rancher/rio/pkg/enterchroot"
	"github.com/sirupsen/logrus"
)

func (a *Agent) Run(ctx *clicontext.CLIContext) error {
	if os.Getuid() != 0 {
		return fmt.Errorf("agent must be ran as root")
	}

	if len(a.T_Token) == 0 {
		return fmt.Errorf("--token is required")
	}

	if len(a.S_Server) == 0 {
		return fmt.Errorf("--server is required")
	}

	dataDir, err := resolvehome.Resolve(a.D_DataDir)
	if err != nil {
		return err
	}

	return RunAgent(a.S_Server, a.T_Token, dataDir, a.L_Log)
}

func RunAgent(server, token, dataDir, logFile string) error {
	dataDir = filepath.Join(dataDir, "agent")

	for {
		tmpFile, err := clientaccess.AgentAccessInfoToTempKubeConfig("", server, token)
		if err != nil {
			logrus.Error(err)
			time.Sleep(2 * time.Second)
			continue
		}
		os.Remove(tmpFile)
		break
	}

	os.Setenv("RIO_URL", server)
	os.Setenv("RIO_TOKEN", token)
	os.Setenv("RIO_DATA_DIR", filepath.Join(dataDir, "root"))

	os.MkdirAll(dataDir, 0700)

	stdout := io.Writer(os.Stdout)
	stderr := io.Writer(os.Stderr)

	if logFile == "" {
		stdout = os.Stdout
		stderr = os.Stderr
	} else {
		l := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    50,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		}
		stdout = l
		stderr = l
	}

	return enterchroot.Mount(dataDir, stdout, stderr)
}

// +build k8s

package server

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"

	"github.com/rancher/rio/pkg/settings"

	"github.com/docker/docker/pkg/reexec"
	"github.com/natefinch/lumberjack"
	"github.com/rancher/norman/signal"
	"github.com/rancher/rio/cli/cmd/agent"
	"github.com/rancher/rio/pkg/server"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func setupLogging(app *cli.Context) {
	if !app.GlobalBool("debug") {
		flag.Set("stderrthreshold", "3")
		flag.Set("alsologtostderr", "false")
		flag.Set("logtostderr", "false")
	}
}

func (s *Server) runWithLogging(app *cli.Context) error {
	l := &lumberjack.Logger{
		Filename:   s.Log,
		MaxSize:    50,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	args := append([]string{"rio"}, os.Args[1:]...)
	cmd := reexec.Command(args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "_RIO_REEXEC_=true")
	cmd.Stderr = l
	cmd.Stdout = l
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (s *Server) Run(app *cli.Context) error {
	if s.Log != "" && os.Getenv("_RIO_REEXEC_") == "" {
		return s.runWithLogging(app)
	}

	if s.ProfilePort > 0 {
		// enable profiler
		go func() {
			log.Fatal(http.ListenAndServe(fmt.Sprintf("localhost:%d", s.ProfilePort), nil))
		}()
	}

	settings.DefaultHTTPOpenPort.Set(s.IngressHttpPort)
	settings.DefaultHTTPSOpenPort.Set(s.IngressHttpsPort)

	setupLogging(app)

	if !s.DisableAgent && os.Getuid() != 0 {
		return fmt.Errorf("must run as root unless --disable-agent is specified")
	}

	logrus.Info("Starting Rio ", app.App.Version)
	ctx := signal.SigTermCancelContext(context.Background())
	sc, err := server.StartServer(ctx, s.D_DataDir, s.L_HttpListenPort, s.P_HttpsListenPort, s.AdvertiseServerIP, !s.DisableControllers, false)
	if err != nil {
		return err
	}

	if s.DisableAgent {
		<-ctx.Done()
		return nil
	}

	dataDir := s.D_DataDir
	if dataDir == "" {
		dataDir = "/var/lib/rancher/rio"
	}

	logFile := filepath.Join(dataDir, "agent/agent.log")
	url := fmt.Sprintf("https://localhost:%d", s.P_HttpsListenPort)
	logrus.Infof("Agent starting, logging to %s", logFile)
	return agent.RunAgent(url, server.FormatToken(sc.NodeToken), dataDir, logFile, s.I_NodeIp)
}

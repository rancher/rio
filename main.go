//go:generate go run types/codegen/cleanup/main.go
//go:generate go run vendor/github.com/jteeuwen/go-bindata/go-bindata/AppendSliceValue.go vendor/github.com/jteeuwen/go-bindata/go-bindata/main.go vendor/github.com/jteeuwen/go-bindata/go-bindata/version.go -o ./stacks/bindata.go -ignore bindata.go -pkg stacks ./stacks/
//go:generate go fmt stacks/bindata.go
//go:generate go run types/codegen/main.go

package main

import (
	"context"
	"net/http"
	"os"

	_ "net/http/pprof"

	"github.com/docker/docker/pkg/reexec"
	"github.com/rancher/norman/signal"
	_ "github.com/rancher/rio/pkg/kubectl"
	"github.com/rancher/rio/pkg/server"
	"github.com/sirupsen/logrus"
)

func main() {
	if reexec.Init() {
		return
	}

	if err := run(); err != nil {
		logrus.Fatal(err)
	}
}

func run() error {
	go func() {
		logrus.Fatal(http.ListenAndServe("localhost:6061", nil))
	}()

	if os.Getenv("RIO_DEBUG") == "true" {
		logrus.SetLevel(logrus.DebugLevel)
	}
	inCluster := os.Getenv("RIO_IN_CLUSTER") == "true"
	ctx := signal.SigTermCancelContext(context.Background())
	_, err := server.StartServer(ctx, "./data-dir", 5080, 5443, "127.0.0.1", true, inCluster)
	<-ctx.Done()
	return err
}

package server

import (
	"context"
	"crypto/sha256"
	cryptotls "crypto/tls"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rancher/norman"
	"github.com/rancher/norman/pkg/resolvehome"
	"github.com/rancher/norman/signal"
	"github.com/rancher/norman/types"
	"github.com/rancher/rancher/pkg/settings"
	"github.com/rancher/rio/api/setup"
	"github.com/rancher/rio/controllers"
	rTypes "github.com/rancher/rio/types"
	"github.com/rancher/rio/types/apis/apiextensions.k8s.io/v1beta1"
	cmv1alpha1 "github.com/rancher/rio/types/apis/certmanager.k8s.io/v1alpha1"
	"github.com/rancher/rio/types/apis/networking.istio.io/v1alpha3"
	policyv1beta1 "github.com/rancher/rio/types/apis/policy/v1beta1"
	projectv1 "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1"
	projectschema "github.com/rancher/rio/types/apis/project.rio.cattle.io/v1/schema"
	autoscalev1 "github.com/rancher/rio/types/apis/rio-autoscale.cattle.io/v1"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	rioschema "github.com/rancher/rio/types/apis/rio.cattle.io/v1/schema"
	storagev1 "github.com/rancher/rio/types/apis/storage.k8s.io/v1"
	projectclient "github.com/rancher/rio/types/client/project/v1"
	"github.com/rancher/rio/types/client/rio/v1"
	"github.com/rancher/types/apis/apps/v1beta2"
	"github.com/rancher/types/apis/core/v1"
	"github.com/rancher/types/apis/management.cattle.io/v3"
	rbacv1 "github.com/rancher/types/apis/rbac.authorization.k8s.io/v1"
	"github.com/sirupsen/logrus"
	net2 "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/kubernetes/pkg/wrapper/server"
)

func NewConfig(dataDir string, inCluster bool) (*norman.Config, error) {
	dataDir, err := resolveDataDir(dataDir)
	if err != nil {
		return nil, err
	}

	return &norman.Config{
		Name: "rio",
		Schemas: []*types.Schemas{
			rioschema.Schemas,
			projectschema.Schemas,
		},

		CRDs: map[*types.APIVersion][]string{
			&rioschema.Version: {
				client.ServiceType,
				client.ConfigType,
				client.RouteSetType,
				client.VolumeType,
				client.StackType,
				client.ExternalServiceType,
			},
			&projectschema.Version: {
				projectclient.ListenConfigType,
				projectclient.PublicDomainType,
				projectclient.FeatureType,
				projectclient.SettingType,
			},
		},

		Clients: []norman.ClientFactory{
			autoscalev1.Factory,
			cmv1alpha1.Factory,
			policyv1beta1.Factory,
			projectv1.Factory,
			rbacv1.Factory,
			riov1.Factory,
			storagev1.Factory,
			v1alpha3.Factory,
			v1beta1.Factory,
			v1beta2.Factory,
			v1.Factory,
			v3.Factory,
		},

		CustomizeSchemas: setup.Types,

		GlobalSetup: rTypes.BuildContext,

		MasterSetup: func(ctx context.Context) (context.Context, error) {
			rTypes.From(ctx).InCluster = inCluster
			return ctx, nil
		},

		MasterControllers: []norman.ControllerRegister{
			rTypes.Register(controllers.Register),
		},

		K3s: norman.K3sConfig{
			DataDir:                dataDir,
			RemoteDialerAuthorizer: authorizer,
		},
	}, nil
}

func resolveDataDir(dataDir string) (string, error) {
	if dataDir == "" {
		if os.Getuid() == 0 {
			dataDir = "/var/lib/rancher/rio"
		} else {
			dataDir = "${HOME}/.rancher/rio"
		}
	}

	dataDir = filepath.Join(dataDir, "server")
	return resolvehome.Resolve(dataDir)
}

func StartServer(ctx context.Context, dataDir string, httpPort, httpsPort int, controllers, inCluster bool) (*server.ServerConfig, error) {
	ctx = signal.SigTermCancelContext(ctx)

	config, err := NewConfig(dataDir, inCluster)
	if err != nil {
		return nil, err
	}

	dataDir = config.K3s.DataDir

	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, errors.Wrapf(err, "can not mkdir %s", dataDir)
	}

	if err := os.Chdir(dataDir); err != nil {
		return nil, errors.Wrapf(err, "can not chdir %s", dataDir)
	}

	ctx, srv, err := config.Build(ctx, &norman.Options{
		DisableControllers: !controllers,
	})
	if err != nil {
		return nil, err
	}

	sc, _ := srv.Runtime.K3sServerConfig.(*server.ServerConfig)

	root := router(sc,
		srv.Runtime.APIHandler,
		srv.Runtime.K3sTunnelServer)

	if err := startTLS(ctx, httpPort, httpsPort, root); err != nil {
		return nil, err
	}

	if sc == nil {
		return nil, nil
	}

	var (
		clientFile string
		nodeFile   string
	)

	if len(sc.ClientToken) > 0 {
		p := filepath.Join(sc.DataDir, "client-token")
		if err := writeToken(sc.ClientToken, p); err != nil {
			return nil, err
		}
		logrus.Infof("Client token is available at %s", p)
		clientFile = p
	}

	if len(sc.NodeToken) > 0 {
		p := filepath.Join(sc.DataDir, "node-token")
		if err := writeToken(sc.NodeToken, p); err != nil {
			return nil, err
		}
		logrus.Infof("Node token is available at %s", p)
		nodeFile = p
	}

	ioutil.WriteFile(filepath.Join(dataDir, "port"), []byte(strconv.Itoa(httpsPort)), 0600)

	if len(clientFile) > 0 {
		printToken(httpsPort, "To use CLI:", clientFile, "login")
	}

	if len(nodeFile) > 0 {
		printToken(httpsPort, "To join node to cluster:", nodeFile, "agent")
	}

	if err := waitForGood(ctx, httpsPort); err != nil {
		return nil, err
	}

	return sc, nil
}

func waitForGood(ctx context.Context, httpsPort int) error {
	rt := &http.Transport{
		TLSClientConfig: &cryptotls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := http.Client{
		Transport: rt,
	}
	defer rt.CloseIdleConnections()

	for {
		time.Sleep(500 * time.Millisecond)

		select {
		case <-ctx.Done():
			return fmt.Errorf("start interrupted")
		default:
		}

		resp, err := client.Get(fmt.Sprintf("https://localhost:%d/healthz", httpsPort))
		if err != nil {
			logrus.Debugf("Waiting for server start: %v", err)
			continue
		}
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			logrus.Debugf("Waiting for server start, read failed: %v", err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return nil
		}

		logrus.Debugf("Waiting for server, non-200 response: %s", bytes)
	}
}

func printToken(httpsPort int, prefix, file, cmd string) error {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		logrus.Error(err)
		return err
	}

	token := strings.TrimSpace(string(content))
	ip, err := net2.ChooseHostInterface()
	if err != nil {
		logrus.Error(err)
		return err
	}

	logrus.Infof("%s rio %s -s https://%s:%d -t %s", prefix, cmd, ip.String(), httpsPort, token)
	return nil
}

func FormatToken(token string) string {
	if len(token) == 0 {
		return token
	}

	prefix := "R10"
	certs := settings.CACerts.Get()
	if len(certs) > 0 {
		digest := sha256.Sum256([]byte(certs))
		prefix = "R10" + hex.EncodeToString(digest[:]) + "::"
	}

	return prefix + token
}

func writeToken(token, file string) error {
	if len(token) == 0 {
		return nil
	}

	token = FormatToken(token)
	return ioutil.WriteFile(file, []byte(token+"\n"), 0600)
}

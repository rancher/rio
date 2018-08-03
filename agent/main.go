package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/coreos/flannel"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/rancher/norman/signal"
	proxy2 "github.com/rancher/rancher/pkg/clusterrouter/proxy"
	"github.com/rancher/rancher/pkg/remotedialer"
	"github.com/rancher/rio/agent/containerd"
	"github.com/rancher/rio/pkg/clientaccess"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/cmd/agent"
)

var (
	ports = map[string]bool{
		"10250": true,
		"10010": true,
	}
)

type AgentConfig struct {
	LocalVolumeDir string
	Config         *agent.AgentConfig
	CACerts        []byte
	TargetHost     string
	Certificate    *tls.Certificate
}

func main() {
	if err := run(); err != nil {
		logrus.Fatal(err)
	}
}

func run() error {
	ctx := signal.SigTermCancelContext(context.Background())

	localURL, err := url.Parse("https://127.0.0.1:6444")
	if err != nil {
		panic(err)
	}

	containerd.Run()

	var agentConfig *AgentConfig
	for {
		agentConfig, err = getConfig(localURL)
		if err != nil {
			logrus.Error(err)
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	if err := runTunnel(agentConfig); err != nil {
		return err
	}

	if err := runProxy(agentConfig); err != nil {
		return err
	}

	if err := agent.Agent(agentConfig.Config); err != nil {
		return err
	}

	if err := waitForNode(agentConfig); err != nil {
		return err
	}

	if err := runFlannel(agentConfig.Config); err != nil {
		return err
	}

	//if err := runLocalStorage(agentConfig); err != nil {
	//	return err
	//}

	<-ctx.Done()
	return nil
}

//func runLocalStorage(config *AgentConfig) error {
//	os.Setenv("KUBECONFIG", config.Config.KubeConfig)
//
//	provisionerConfig := common.ProvisionerConfiguration{
//		StorageClassConfig: map[string]common.MountConfig{
//			"local": {
//				HostDir:  config.LocalVolumeDir,
//				MountDir: config.LocalVolumeDir,
//			},
//		},
//		MinResyncPeriod: metav1.Duration{Duration: 5 * time.Minute},
//	}
//
//	nodeName := config.Config.NodeName
//
//	client := common.SetupClient()
//	node := getNode(client, nodeName)
//
//	glog.Info("Starting controller\n")
//	procTable := deleter.NewProcTable()
//	go controller.StartLocalController(client, procTable, &common.UserConfig{
//		Node:              node,
//		DiscoveryMap:      provisionerConfig.StorageClassConfig,
//		NodeLabelsForPV:   provisionerConfig.NodeLabelsForPV,
//		UseAlphaAPI:       provisionerConfig.UseAlphaAPI,
//		UseJobForCleaning: provisionerConfig.UseJobForCleaning,
//		MinResyncPeriod:   provisionerConfig.MinResyncPeriod,
//	})
//
//	return nil
//}

func waitForNode(config *AgentConfig) error {
	os.Setenv("KUBECONFIG", config.Config.KubeConfig)

	nodeName := config.Config.NodeName

	restConfig, err := clientcmd.BuildConfigFromFlags("", config.Config.KubeConfig)
	if err != nil {
		return err
	}

	client, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return err
	}

	for {
		node, err := client.CoreV1().Nodes().Get(nodeName, metav1.GetOptions{})
		if err == nil && node.Spec.PodCIDR != "" {
			return nil
		}
		if err == nil {
			logrus.Infof("waiting for node %s CIDR not assigned yet", nodeName)
		} else {
			logrus.Infof("waiting for node %s: %v", nodeName, err)
		}
		time.Sleep(2 * time.Second)
	}
}

func runTunnel(config *AgentConfig) error {
	restConfig, err := clientcmd.BuildConfigFromFlags("", config.Config.KubeConfig)
	if err != nil {
		return err
	}

	transportConfig, err := restConfig.TransportConfig()
	if err != nil {
		return err
	}

	wsURL := fmt.Sprintf("wss://%s/v1beta1/connect", config.TargetHost)
	headers := map[string][]string{
		"X-Rio-NodeName": {config.Config.NodeName},
	}
	ws := &websocket.Dialer{}

	if len(config.CACerts) > 0 {
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(config.CACerts)
		ws.TLSClientConfig = &tls.Config{
			RootCAs: pool,
		}
	}

	if transportConfig.Username != "" {
		auth := transportConfig.Username + ":" + transportConfig.Password
		auth = base64.StdEncoding.EncodeToString([]byte(auth))
		headers["Authorization"] = []string{"Basic " + auth}
	}

	once := sync.Once{}
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for {
			logrus.Infof("Connecting to %s", wsURL)
			remotedialer.ClientConnect(wsURL, http.Header(headers), ws, func(proto, address string) bool {
				host, port, err := net.SplitHostPort(address)
				return err == nil && proto == "tcp" && ports[port] && host == "127.0.0.1"
			}, func(_ context.Context) error {
				once.Do(wg.Done)
				return nil
			})
			time.Sleep(5 * time.Second)
		}
	}()

	wg.Wait()
	return nil
}

func runProxy(config *AgentConfig) error {
	proxy, err := proxy2.NewSimpleProxy(config.TargetHost, config.CACerts)
	if err != nil {
		return err
	}

	listener, err := tls.Listen("tcp", "127.0.0.1:6444", &tls.Config{
		Certificates: []tls.Certificate{
			*config.Certificate,
		},
	})

	if err != nil {
		return errors.Wrap(err, "Failed to start tls listener")
	}

	go func() {
		err := http.Serve(listener, proxy)
		logrus.Fatalf("TLS proxy stopped: %v", err)
	}()

	return nil
}

func runFlannel(config *agent.AgentConfig) error {
	go func() {
		flannel.Main([]string{
			"--ip-masq",
			"--kubeconfig-file", config.KubeConfig,
		})

		logrus.Fatalf("flannel exited")
	}()
	return nil
}

func getConfig(localURL *url.URL) (*AgentConfig, error) {
	u := os.Getenv("RIO_URL")
	if u == "" {
		return nil, fmt.Errorf("RIO_URL env var is required")
	}

	uParsed, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("RIO_URL [%s] is invalid: %v", u, err)
	}

	t := os.Getenv("RIO_TOKEN")
	if t == "" {
		return nil, fmt.Errorf("RIO_TOKEN env var is required")
	}

	dataDir := os.Getenv("RIO_DATA_DIR")
	if dataDir == "" {
		return nil, fmt.Errorf("RIO_DATA_DIR is required")
	}
	os.MkdirAll(dataDir, 0700)

	kubeConfig := filepath.Join(dataDir, "kubeconfig.yaml")

	_, cidr, _ := net.ParseCIDR("10.42.0.0/16")

	cacerts, tlsCert, err := clientaccess.AgentAccessInfoToKubeConfig(kubeConfig, u, t, localURL)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get access info")
	}

	clientCABytes := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: tlsCert.Certificate[1],
	})
	clientCA := filepath.Join(dataDir, "client-ca.pem")
	if err := ioutil.WriteFile(clientCA, clientCABytes, 0600); err != nil {
		return nil, errors.Wrapf(err, "failed to write client CA")
	}

	nodeName, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	nodeName = strings.Split(nodeName, ".")[0]

	return &AgentConfig{
		LocalVolumeDir: filepath.Join(dataDir, "local"),
		Config: &agent.AgentConfig{
			NodeName:      nodeName,
			ClusterCIDR:   *cidr,
			KubeConfig:    kubeConfig,
			RuntimeSocket: "/run/rio/containerd.sock",
			CNIBinDir:     "/usr/share/cni",
			CACertPath:    clientCA,
			ListenAddress: "127.0.0.1",
		},
		CACerts:     cacerts,
		TargetHost:  uParsed.Host,
		Certificate: tlsCert,
	}, err
}

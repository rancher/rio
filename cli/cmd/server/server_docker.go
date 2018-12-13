// +build darwin windows

package server

import (
	"bufio"
	"context"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/rancher/norman/signal"
	"github.com/rancher/norman/types/slice"
	"github.com/rancher/rio/cli/cmd/login"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/pkg/clientaccess"
	"github.com/rancher/rio/pkg/name"
	"github.com/rancher/rio/pkg/settings"
	"github.com/rancher/rio/version"
	"github.com/urfave/cli"
	"k8s.io/client-go/util/homedir"
)

const (
	rioNameBase = "rio"
	imageBase   = "daishan1992/rio:"
	localhost   = "127.0.0.1"
)

var rioForMac = `

|  _ \(_) ___   |  ___|__  _ __  |  \/  | __ _  ___ 
| |_) | |/ _ \  | |_ / _ \| '__| | |\/| |/ _' |/ __|
|  _ <| | (_) | |  _| (_) | |    | |  | | (_| | (__
|_| \_\_|\___/  |_|  \___/|_|    |_|  |_|\__,_|\___|

Rio for Mac
`

var rioForWindows = `

|  _ \(_) ___   |  ___|__  _ __  \ \      / (_)_ __   __| | _____      _____ 
| |_) | |/ _ \  | |_ / _ \| '__|  \ \ /\ / /| | '_ \ / _' |/ _ \ \ /\ / / __|
|  _ <| | (_) | |  _| (_) | |      \ V  V / | | | | | (_| | (_) \ V  V /\__ \
|_| \_\_|\___/  |_|  \___/|_|       \_/\_/  |_|_| |_|\__,_|\___/ \_/\_/ |___/

Rio For Windows
`

type Server struct {
	P_HttpsListenPort string `desc:"HTTPS listen port" default:"8443"`
	L_HttpListenPort  string `desc:"HTTP listen port" default:"8080"`
	IngressHttpPort   string `desc:"Ingress HTTP listen port" default:"80"`
	IngressHttpsPort  string `desc:"Ingress HTTP listen port" default:"443"`
}

func (s *Server) Customize(command *cli.Command) {
	command.Category = "CLUSTER RUNTIME"
}

func (s *Server) Run(appCtx *clicontext.CLIContext) error {
	ctx := signal.SigTermCancelContext(context.Background())
	dc, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	// try to find existing container
	var cont *types.Container
	containerList, err := dc.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return err
	}
	for _, c := range containerList {
		if slice.ContainsString(c.Names, "/"+rioNameBase) {
			cont = &c
			break
		}
	}

	// generate admin token
	home := filepath.Join(homedir.HomeDir(), ".rancher", "rio")
	if err := os.MkdirAll(home, 0755); err != nil {
		return err
	}
	tokenFile := filepath.Join(home, "admin-token")
	if _, err := os.Stat(tokenFile); err != nil {
		token := make([]byte, 16, 16)
		_, err := cryptorand.Read(token)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(tokenFile, []byte(hex.EncodeToString(token)), 0755); err != nil {
			return err
		}
	}
	token, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return err
	}

	// if container doesn't exist then create new container
	if cont == nil {
		if err := startNewContainer(ctx, dc, string(token), s.L_HttpListenPort, s.P_HttpsListenPort, s.IngressHttpPort, s.IngressHttpsPort); err != nil {
			return err
		}
	}

	// if container exist and should be upgraded, upgrade rio container
	if cont != nil {
		if shouldUpgrade(cont, s.L_HttpListenPort, s.P_HttpsListenPort, s.IngressHttpPort, s.IngressHttpsPort) {
			// stop and remove existing container
			if err := deleteOldContainer(ctx, dc, cont.ID); err != nil {
				return err
			}

			//create new container
			if err := startNewContainer(ctx, dc, string(token), s.L_HttpListenPort, s.P_HttpsListenPort, s.IngressHttpPort, s.IngressHttpsPort); err != nil {
				return err
			}
		}
	}

	fmt.Println("Rio server is running. To view logs, run `docker logs -f rio`")
	fmt.Println("Automatically logging in right now...")

	serverUrl := "https://127.0.0.1:" + s.P_HttpsListenPort
	var loginErr error
	login := false
	for i := 0; i < 10; i++ {
		err := loginRio(appCtx, serverUrl, string(token))
		if err == nil {
			login = true
			break
		}
		loginErr = err
		time.Sleep(time.Second * 2)
	}
	if !login {
		return errors.Errorf("Login failed. Error: %v. Try to re-run `rio server`", loginErr)
	}
	fmt.Println("Log in successful. Welcome to Rio!")
	if runtime.GOOS == "darwin" {
		fmt.Print(rioForMac)
	} else {
		fmt.Print(rioForWindows)
	}
	return nil
}

func deleteOldContainer(ctx context.Context, dc *client.Client, id string) error {
	if err := dc.ContainerStop(ctx, id, nil); err != nil {
		return err
	}
	return dc.ContainerRemove(ctx, id, types.ContainerRemoveOptions{})
}

func startNewContainer(ctx context.Context, dc *client.Client, token, httpPort, httpsPort, ingressHttpPort, ingressHttpsPort string) error {
	// pull image
	if err := pullImage(dc, imageBase+version.Version); err != nil {
		return err
	}

	// create and start container
	config, hostConfig, err := constructRioContainer(httpPort, httpsPort, ingressHttpPort, ingressHttpsPort, token)
	if err != nil {
		return err
	}
	newContainer, err := dc.ContainerCreate(ctx, config, hostConfig, nil, rioNameBase)
	if err != nil {
		return err
	}

	return dc.ContainerStart(ctx, newContainer.ID, types.ContainerStartOptions{})
}

func shouldUpgrade(cont *types.Container, httpPort, httpsPort, ingressHttpPort, ingressHttpsPort string) bool {
	if cont.Image != imageBase+version.Version {
		return true
	}
	m := map[string]bool{}
	for _, p := range cont.Ports {
		m[strconv.Itoa(int(p.PrivatePort))] = true
	}
	return !m[httpPort] || !m[httpsPort] || !m[ingressHttpPort] || !m[ingressHttpsPort]
}

func loginRio(ctx *clicontext.CLIContext, serverUrl, adminToken string) error {
	url, err := url.Parse(serverUrl)
	if err != nil {
		return errors.Wrapf(err, "Failed to parse RIO_URL %s", serverUrl)
	}
	cacerts, err := clientaccess.GetCACerts(*url)
	if err != nil {
		return errors.Wrapf(err, "Failed to get cacerts from %s", serverUrl)
	}
	digest := sha256.Sum256([]byte(cacerts))
	newHash := hex.EncodeToString(digest[:])
	token := fmt.Sprintf("R10%s::admin:%s", newHash, adminToken)
	cluster, err := login.Validate(serverUrl, token)
	if err != nil {
		return err
	}

	cluster.ID = name.Hex(cluster.URL, 5)
	cluster.Name = cluster.ID
	return ctx.Config.SaveCluster(cluster, true)
}

func constructRioContainer(httpPort, httpsPort, ingressHttpPort, ingressHttpsPort, token string) (*container.Config, *container.HostConfig, error) {
	rioHttps, err := nat.NewPort("tcp", httpsPort)
	if err != nil {
		return nil, nil, err
	}
	gatewayHttp, err := nat.NewPort("tcp", settings.DefaultHTTPOpenPort.Get())
	if err != nil {
		return nil, nil, err
	}
	gatewayHttps, err := nat.NewPort("tcp", settings.DefaultHTTPSOpenPort.Get())
	if err != nil {
		return nil, nil, err
	}

	config := &container.Config{
		Image: imageBase + version.Version,
		Env: []string{
			fmt.Sprintf("RIO_SERVER_IP=%s", localhost),
			fmt.Sprintf("K3S_ADMIN_TOKEN=%s", token),
			fmt.Sprintf("ADVERTISE_RDNS_CLUSTER_IP=%s", localhost),
		},
		OpenStdin: true,
		Tty:       true,
		Volumes: map[string]struct{}{
			"rio-data": {},
		},
		ExposedPorts: map[nat.Port]struct{}{
			rioHttps:     {},
			gatewayHttp:  {},
			gatewayHttps: {},
		},
		Cmd: []string{
			"/rio",
			"server",
			"--https-listen-port",
			httpsPort,
			"--http-listen-port",
			httpPort,
			"--ingress-http-port",
			ingressHttpPort,
			"--ingress-https-port",
			ingressHttpsPort,
		},
		Hostname: rioNameBase,
	}

	hostConfig := &container.HostConfig{
		Binds: []string{
			"rio-data:/var/lib/rancher/rio",
		},
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
		PortBindings: map[nat.Port][]nat.PortBinding{
			rioHttps: {
				{
					HostPort: httpsPort,
				},
			},
			gatewayHttp: {
				{
					HostPort: settings.DefaultHTTPOpenPort.Get(),
				},
			},
			gatewayHttps: {
				{
					HostPort: settings.DefaultHTTPSOpenPort.Get(),
				},
			},
		},
		Privileged: true,
		Tmpfs: map[string]string{
			"/run":     "",
			"/var/run": "",
		},
	}

	return config, hostConfig, nil
}

func pullImage(dc *client.Client, image string) error {
	reader, err := dc.ImagePull(context.Background(), image, types.ImagePullOptions{})
	if err != nil {
		return errors.Wrap(err, "Failed to pull image")
	}
	defer reader.Close()
	return wrapReader(reader, image)
}

func wrapReader(reader io.ReadCloser, imageUUID string) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		status, err := toMap(scanner.Text())
		if err != nil {
			return err
		}
		if hasKey(status, "error") {
			return fmt.Errorf("image [%s] failed to pull: %v", imageUUID, status["error"])
		}
	}
	return nil
}

func toMap(rawstring string) (map[string]interface{}, error) {
	obj := map[string]interface{}{}
	err := json.Unmarshal([]byte(rawstring), &obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func hasKey(m interface{}, key string) bool {
	_, ok := m.(map[string]interface{})[key]
	return ok
}

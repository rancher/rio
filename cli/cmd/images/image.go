package images

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/rancher/rio/cli/pkg/build"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/localbuilder"
	"github.com/rancher/rio/cli/pkg/tables"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/wrangler/pkg/kv"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	portForward    = "9998"
	registryPort   = "80"
	registryAPIURL = "http://127.0.0.1"
)

type catalog struct {
	Repositories []string `json:"repositories,omitempty"`
}

type repository struct {
	Name string   `json:"name,omitempty"`
	Tags []string `json:"tags,omitempty"`
}

type Image struct {
	Namespace string
	Repo      string
	Tag       string
	Image     string
}

type Images struct {
}

func (i *Images) Customize(cmd *cli.Command) {
	cmd.ShortName = "image"
}

func (i *Images) Run(ctx *clicontext.CLIContext) error {
	if err := build.EnableBuildAndWait(ctx); err != nil {
		logrus.Warn(err)
	}

	pods, err := ctx.Core.Pods(ctx.SystemNamespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, constants.BuildkitdService) {
			readyChan := make(chan struct{})
			go func() {
				if err := localbuilder.PortForward(ctx.K8s, portForward, registryPort, pod, false, readyChan, localbuilder.ChanWrapper(ctx.Ctx.Done())); err != nil {
					logrus.Fatal(err)
				}
			}()

			select {
			case <-readyChan:
				break
			}

			var result []Image
			resp, err := http.Get(fmt.Sprintf("%s:%s/v2/_catalog", registryAPIURL, portForward))
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}

			catalog := catalog{}
			if err := json.Unmarshal(data, &catalog); err != nil {
				return err
			}

			for _, repo := range catalog.Repositories {
				resp, err = http.Get(fmt.Sprintf("%s:%s/v2/%s/tags/list", registryAPIURL, portForward, repo))
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				data, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return err
				}

				repo := repository{}
				if err := json.Unmarshal(data, &repo); err != nil {
					return err
				}

				for _, tag := range repo.Tags {
					namespace, name := kv.Split(repo.Name, "/")
					if namespace != ctx.GetSetNamespace() {
						continue
					}
					result = append(result, Image{
						Namespace: namespace,
						Repo:      name,
						Tag:       tag,
						Image:     fmt.Sprintf("localhost:5442/%s:%s", repo.Name, tag),
					})
				}
			}

			writer := tables.NewImage(ctx)
			defer writer.TableWriter().Close()
			for _, obj := range result {
				writer.TableWriter().Write(obj)
			}
			return writer.TableWriter().Err()
		}
	}

	return fmt.Errorf("buildkitd pod not found")
}

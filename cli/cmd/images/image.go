package images

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/localbuilder"
	"github.com/rancher/rio/cli/pkg/tables"
	"github.com/rancher/rio/pkg/constants"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type catalog struct {
	Repositories []string `json:"repositories,omitempty"`
}

type repository struct {
	Name string   `json:"name,omitempty"`
	Tags []string `json:"tags,omitempty"`
}

type Image struct {
	Repo  string
	Tag   string
	Image string
}

type Images struct {
}

func (i *Images) Run(ctx *clicontext.CLIContext) error {
	pods, err := ctx.Core.Pods(ctx.SystemNamespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, constants.BuildkitdService) {
			readyChan := make(chan struct{})
			go func() {
				if err := localbuilder.PortForward(ctx.K8s, "9998", "80", pod, false, readyChan, localbuilder.ChanWrapper(ctx.Ctx.Done())); err != nil {
					logrus.Fatal(err)
				}
			}()

			select {
			case <-readyChan:
				break
			}

			var result []Image
			resp, err := http.Get("http://127.0.0.1:9998/v2/_catalog")
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
				resp, err = http.Get(fmt.Sprintf("http://127.0.0.1:9998/v2/%s/tags/list", repo))
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
					result = append(result, Image{
						Repo:  repo.Name,
						Tag:   tag,
						Image: fmt.Sprintf("localhost:5442/%s:%s", repo.Name, tag),
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

package up

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rancher/rio/cli/cmd/apply"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/localbuilder"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/stack"
)

const (
	defaultRiofile       = "Riofile"
	defaultRiofileAnswer = "Riofile-answers"

	defaultRiofileContent = `
services:
  %s:
    image: ./
    ports: 80:8080/http`
)

func Build(builds map[stack.ContainerBuildKey]riov1.ImageBuildSpec, c *clicontext.CLIContext, parallel bool) (map[stack.ContainerBuildKey]string, error) {
	if len(builds) == 0 {
		return nil, nil
	}
	localBuilder, err := localbuilder.NewLocalBuilder(c.Ctx, c.SystemNamespace, c.Apply, c.K8s)
	if err != nil {
		return nil, err
	}

	images, err := localBuilder.Build(c.Ctx, builds, parallel, c.GetSetNamespace())
	if err != nil {
		return nil, err
	}
	for k, i := range images {
		if strings.HasPrefix(i, constants.RegistryService) {
			images[k] = strings.Replace(i, constants.RegistryService, constants.LocalRegistry, -1)
		}
	}

	return images, nil
}

func GetCurrentDir() string {
	workingDir, _ := os.Getwd()
	dir := filepath.Base(workingDir)
	return strings.ToLower(dir)
}

func LoadRiofile(path string) ([]byte, error) {
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			return ioutil.ReadFile(path)
		}
	}

	if _, err := os.Stat(defaultRiofile); err == nil {
		return ioutil.ReadFile(defaultRiofile)
	}

	return []byte(fmt.Sprintf(defaultRiofileContent, GetCurrentDir())), nil
}

func readFile(file string) ([]byte, error) {
	if file == "-" {
		return ioutil.ReadAll(os.Stdin)
	}
	if strings.HasPrefix(file, "http") {
		resp, err := http.Get(file)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)
	}
	return ioutil.ReadFile(file)
}

func LoadAnswer(path string) (map[string]string, error) {
	if path == "" {
		if _, err := os.Stat(defaultRiofileAnswer); err == nil {
			path = defaultRiofileAnswer
		}
	}
	return apply.ReadAnswers(path)
}

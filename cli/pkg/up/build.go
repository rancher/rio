package up

import (
	"bytes"
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
	"sigs.k8s.io/yaml"
)

const (
	defaultRiofile       = "Riofile"
	defaultRiofileAnswer = "Riofile-answers"

	defaultRiofileContent = `
services:
  %s:
    image: ./
    ports: 80:8080/http`

	defaultRiofileContentWithDockerfile = `
services:
  %s:
    build:
      dockerfile: %s
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

// LoadRiofile handles the following scenarios:
// An assumed Riofile: rio up
// An assumed Dockerfile: rio up
// A named Riofile: rio up -f myRiofile
// A named Dockerfile: rio up -f myDockerfile
func LoadRiofile(path string) ([]byte, error) {
	if path != "" {
		content, err := readFile(path)
		if err != nil {
			return nil, err
		}
		// named Riofile, has either valid yaml or templating
		var r map[string]interface{}
		if err := yaml.Unmarshal(content, &r); err == nil || bytes.Contains(content, []byte("goTemplate:")) {
			return content, nil
		}
		// named Dockerfile
		return []byte(fmt.Sprintf(defaultRiofileContentWithDockerfile, GetCurrentDir(), path)), nil
	}
	// assumed Riofile
	if _, err := os.Stat(defaultRiofile); err == nil {
		return ioutil.ReadFile(defaultRiofile)
	}
	// assumed Dockerfile
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

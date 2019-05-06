package up

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/docker/docker/pkg/symlink"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/pkg/systemstack"
)

func readFileInPath(relativePath, file string) ([]byte, error) {
	if strings.HasPrefix(relativePath, "http") {
		base, err := url.Parse(relativePath)
		if err != nil {
			return nil, err
		}
		ref, err := url.Parse(file)
		if err != nil {
			return nil, err
		}
		resolved := base.ResolveReference(ref)
		if err != nil {
			return nil, err
		}
		resp, err := http.Get(resolved.String())
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)
	}
	f, err := symlink.FollowSymlinkInScope(file, relativePath)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadFile(f)
}

func Run(ctx *clicontext.CLIContext, content []byte, namespace string, answers map[string]string) error {
	stack := systemstack.NewStack(ctx.Apply, namespace, namespace, true)
	stack.WithContent(content)

	return stack.Deploy(answers)
}

package util

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/docker/app/pkg/resto"
)

const StackFileKey = "rio-stack.yaml"

// ReadFile reads a file reference from different sources and return a map of files which contains stack definition file and other configs
// A file reference can be generated from OS stand input, http uri, local disk or docker image.
func ReadFile(file string) (map[string]string, error) {
	// stdin
	if file == "-" {
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, err
		}
		return map[string]string{
			StackFileKey: string(data),
		}, nil
	}

	// http uri
	if strings.HasPrefix(file, "http") {
		resp, err := http.Get(file)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return map[string]string{
			StackFileKey: string(data),
		}, nil
	}

	// docker registry
	if _, err := os.Stat(file); err != nil {
		files, err := resto.PullConfigMulti(context.Background(), file, resto.RegistryOptions{})
		// ignore the error and fall back to local files
		if err == nil {
			return files, nil
		}
	}

	// local disk
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		StackFileKey: string(data),
	}, nil
}

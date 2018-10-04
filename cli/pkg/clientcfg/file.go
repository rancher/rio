package clientcfg

import (
	"io/ioutil"
	"os"
	"strings"
)

func defaultNameFromFile(file string) (string, error) {
	data, err := ioutil.ReadFile(file)
	if os.IsNotExist(err) {
		return "default", nil
	} else if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

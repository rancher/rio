package util

import (
	"io/ioutil"
	"os"
)

func ReadFile(file string) ([]byte, error) {
	if file == "-" {
		return ioutil.ReadAll(os.Stdin)
	}
	return ioutil.ReadFile(file)
}

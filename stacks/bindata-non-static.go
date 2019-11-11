// +build !static

package stacks

import (
	"io/ioutil"
)

func Asset(name string) ([]byte, error) {
	return ioutil.ReadFile(name)
}

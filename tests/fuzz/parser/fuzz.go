// +build gofuzz

package riofile

import "github.com/rancher/rio/pkg/riofile"
import "github.com/rancher/rio/pkg/template"

func Fuzz(data []byte) int {
	if _, err := riofile.Parse(data, template.AnswersFromMap(nil)); err != nil {
		return 0
	}
	return 1
}

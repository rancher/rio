package template

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

func ReadDir(suffix, dir string) (map[string]*Template, error) {
	files, err := ioutil.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	result := map[string]*Template{}

	for _, f := range files {
		var (
			stackName string
			name      = normalizeYaml(f.Name())
		)

		if f.IsDir() {
			continue
		}

		switch {
		case name == suffix:
			stackName = "default"
		case strings.HasSuffix(name, "-"+suffix):
			stackName = strings.TrimSuffix(name, "-"+suffix)
		default:
			continue
		}

		fName := filepath.Join(dir, f.Name())
		content, err := ioutil.ReadFile(fName)

		if err != nil {
			return nil, errors.Wrapf(err, "reading %s", fName)
		}

		template := &Template{
			Content: content,
		}

		if err := template.Validate(); err != nil {
			return nil, errors.Wrapf(err, "invalid template %s", fName)
		}

		result[stackName] = template
	}

	return result, nil
}

func normalizeYaml(s string) string {
	return strings.Replace(s, ".yml", ".yaml", 1)
}

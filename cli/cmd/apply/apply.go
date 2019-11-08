package apply

import (
	"os"

	"github.com/rancher/norman/pkg/types/convert"
	"github.com/rancher/rio/cli/cmd/util"
	"github.com/rancher/wrangler/pkg/yaml"
)

func ReadAnswers(answersFile string) (map[string]string, error) {
	content, err := util.ReadFile(answersFile)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{}
	if err := yaml.Unmarshal(content, &data); err != nil {
		return nil, err
	}

	result := map[string]string{}
	for k, v := range data {
		result[k] = convert.ToString(v)
	}

	return result, nil
}

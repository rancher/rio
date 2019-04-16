package template

import v1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"

func AnswersFromMap(answers map[string]string) AnswerCallback {
	return func(key string, questions []v1.Question) (string, error) {
		return answers[key], nil
	}
}

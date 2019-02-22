package util

import (
	"github.com/rancher/rio/cli/pkg/clientcfg"
	riov1 "github.com/rancher/rio/types/apis/rio.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func StacksByID(client *clientcfg.KubeClient, projectName string) (map[string]*riov1.Stack, error) {
	stacks, err := client.Rio.Stacks(projectName).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	stackByID := map[string]*riov1.Stack{}
	for i, stack := range stacks.Items {
		stackByID[stack.Name] = &stacks.Items[i]
	}

	return stackByID, nil
}

package util

import (
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	v1 "github.com/rancher/rio/pkg/generated/clientset/versioned/typed/rio.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func StacksByID(client v1.RioV1Interface, projectName string) (map[string]*riov1.Stack, error) {
	stacks, err := client.Stacks(projectName).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	stackByID := map[string]*riov1.Stack{}
	for i, stack := range stacks.Items {
		stackByID[stack.Name] = &stacks.Items[i]
	}

	return stackByID, nil
}

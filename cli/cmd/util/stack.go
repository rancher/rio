package util

import (
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

func StacksByID(wc *client.Client) (map[string]*client.Stack, error) {
	stacks, err := wc.Stack.List(nil)
	if err != nil {
		return nil, err
	}

	stackByID := map[string]*client.Stack{}
	for i, stack := range stacks.Data {
		stackByID[stack.ID] = &stacks.Data[i]
	}

	return stackByID, nil
}

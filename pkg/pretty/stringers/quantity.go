package stringers

import "k8s.io/apimachinery/pkg/api/resource"

func ParseQuantity(num string) (result resource.Quantity, err error) {
	if num == "" {
		return
	}

	return resource.ParseQuantity(num)
}

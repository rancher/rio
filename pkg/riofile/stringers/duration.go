package stringers

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ParseDuration(num string) (result metav1.Duration, err error) {
	if num == "" {
		return
	}

	dur, err := time.ParseDuration(num)
	return metav1.Duration{
		Duration: dur,
	}, err
}

package util

import (
	"fmt"
	"time"

	"github.com/rancher/rio/pkg/settings"
)

func WaitForClusterDomain() error {
	start := time.Millisecond * 250
	for i := 0; i < 5; i++ {
		if settings.ClusterDomain.Get() != "" {
			return nil
		}
		time.Sleep(start)
		start *= 2
	}
	return fmt.Errorf("waiting for cluster domain")
}

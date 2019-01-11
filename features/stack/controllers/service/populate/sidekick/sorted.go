package sidekick

import (
	"sort"

	"github.com/rancher/rio/types/apis/rio.cattle.io/v1"
)

func SortedKeys(m map[string]v1.SidekickConfig) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

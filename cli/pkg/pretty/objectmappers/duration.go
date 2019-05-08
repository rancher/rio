package objectmappers

import (
	"fmt"
	"time"
)

func ParseDurationUnit(s, name string, unit time.Duration) (int, error) {
	if s == "" {
		return 0, nil
	}

	dur, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s: %v", name, err)
	}
	return int(dur / unit), nil
}

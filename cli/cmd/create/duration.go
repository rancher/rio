package create

import (
	"fmt"
	"time"
)

func ParseDurationUnit(s, name string, unit time.Duration) (int64, error) {
	if s == "" {
		return 0, nil
	}

	dur, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s: %v", name, err)
	}
	return int64(dur / unit), nil
}

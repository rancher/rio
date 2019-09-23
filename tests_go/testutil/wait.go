package testutil

import (
	"time"
)

type waitForMe = func() bool

// WaitFor takes a method and waits until it returns true, see WaitForService
func WaitFor(f waitForMe, timeout int) bool {
	sleepSeconds := 1
	for start := time.Now(); time.Since(start) < time.Second*time.Duration(timeout); {
		out := f()
		if out == true {
			return out
		}
		time.Sleep(time.Second * time.Duration(sleepSeconds))
		sleepSeconds++
	}
	return false
}

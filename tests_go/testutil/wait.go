package testutil

import (
	"errors"
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

func WaitForService(name string) error {
	f := func() bool {
		s, err := InspectService(name)
		if err == nil {
			if s.Status.DeploymentStatus != nil && s.Status.DeploymentStatus.AvailableReplicas > 0 {
				return true
			}
		}
		return false
	}
	ok := WaitFor(f, 120)
	if ok == false {
		return errors.New("service failed to initiate")
	}
	return nil
}

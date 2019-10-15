package kubeutils

import (
	"os"

	"github.com/solo-io/go-utils/log"
)

const (
	TERMINATION_LOG = "/dev/termination-log"
)

/*
This function is in kubeutils because it is meant to only be used in containers deployed in kubernetes

Upon termination kubernetes pods will read from a log file in the container and output the contents to
the pod spec for debugging. https://kubernetes.io/docs/tasks/debug-application-cluster/determine-reason-pod-failure/
*/

func LogFailureState(failureErr error) {
	file, err := os.OpenFile(TERMINATION_LOG, os.O_RDWR, 0)
	if err != nil {
		// termination log file does not exist - this can happen in non-kube environments so it is a no-op
		return
	}
	_, err = file.Write([]byte(failureErr.Error()))
	if err != nil {
		// we failed to write to termination log, this should never happen
		log.Fatalf("failed to write error %s due to %s", failureErr.Error(), err.Error())
	}
}

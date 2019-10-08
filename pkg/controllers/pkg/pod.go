package pkg

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
)

func PodDetail(pod *v1.Pod) string {
	output := strings.Builder{}
	if pod == nil {
		return ""
	}
	for _, con := range append(pod.Status.ContainerStatuses, pod.Status.InitContainerStatuses...) {
		if con.State.Waiting != nil && con.State.Waiting.Reason != "" {
			output.WriteString("; ")
			reason := con.State.Waiting.Reason
			if con.State.Waiting.Message != "" {
				reason = reason + "/" + con.State.Waiting.Message
			}
			output.WriteString(fmt.Sprintf("%s(%s)", con.Name, reason))
		}

		if con.State.Terminated != nil && con.State.Terminated.ExitCode != 0 {
			output.WriteString(";")
			if con.State.Terminated.Message == "" {
				con.State.Terminated.Message = "exit code not zero"
			}
			reason := con.State.Terminated.Reason
			if con.State.Terminated.Message != "" {
				reason = reason + "/" + con.State.Terminated.Message
			}
			output.WriteString(fmt.Sprintf("%s(%s), exit code: %v", con.Name, reason, con.State.Terminated.ExitCode))
		}
	}
	return strings.Trim(output.String(), "; ")
}

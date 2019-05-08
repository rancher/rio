package stack

import v1 "k8s.io/api/core/v1"

func MeshEnabled(ns *v1.Namespace) bool {
	return ns.Labels["rio.cattle.io/mesh"] != "false"
}

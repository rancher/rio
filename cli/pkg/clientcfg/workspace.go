package clientcfg

import v1 "k8s.io/api/core/v1"

type Project struct {
	Project *v1.Namespace
	Cluster *Cluster
	Default bool
}

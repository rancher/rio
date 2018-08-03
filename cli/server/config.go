package server

import (
	"github.com/rancher/rio/cli/pkg/resolvehome"
)

const (
	rioConf  = "${HOME}/.rancher/rio/client/rio.yaml"
	k8sConf  = "${HOME}/.rancher/rio/client/k8s.yaml"
	confHome = "${HOME}/.rancher/rio/client"
)

func ConfigHome() (string, error) {
	return resolvehome.Resolve(confHome)
}

func RioConfPath() (string, error) {
	return resolvehome.Resolve(rioConf)
}

func K8sConfPath() (string, error) {
	return resolvehome.Resolve(k8sConf)
}

func Paths() (string, string, string, error) {
	ch, err := ConfigHome()
	if err != nil {
		return "", "", "", err
	}

	rio, err := RioConfPath()
	if err != nil {
		return "", "", "", err
	}

	k8s, err := K8sConfPath()
	if err != nil {
		return "", "", "", err
	}

	return ch, rio, k8s, nil
}

package clientcfg

import (
	projectclient "github.com/rancher/rio/types/client/project/v1"
	"github.com/rancher/rio/types/client/rio/v1"
)

type Project struct {
	projectclient.Project
	Cluster *Cluster
	Default bool

	client *client.Client
}

func (w *Project) Client() (*client.Client, error) {
	if w.client != nil {
		return w.client, nil
	}

	ci, err := w.Cluster.getClientInfo()
	if err != nil {
		return nil, err
	}

	client, err := ci.projectClient(w.ID)
	if err != nil {
		return nil, err
	}

	w.client = client
	return client, err
}

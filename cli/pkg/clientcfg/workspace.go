package clientcfg

import (
	"github.com/rancher/rio/types/client/rio/v1beta1"
	spaceclient "github.com/rancher/rio/types/client/space/v1beta1"
)

type Workspace struct {
	spaceclient.Space
	Cluster *Cluster
	Default bool

	client *client.Client
}

func (w *Workspace) Client() (*client.Client, error) {
	if w.client != nil {
		return w.client, nil
	}

	ci, err := w.Cluster.getClientInfo()
	if err != nil {
		return nil, err
	}

	client, err := ci.workspaceClient(w.ID)
	if err != nil {
		return nil, err
	}

	w.client = client
	return client, err
}

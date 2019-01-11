package objectset

import (
	"github.com/rancher/norman/controller"
	"github.com/rancher/norman/objectclient"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Client interface {
	Generic() controller.GenericController
	ObjectClient() *objectclient.ObjectClient
}

type Processor struct {
	setID       string
	codeVersion string
	clients     map[schema.GroupVersionKind]Client
}

func NewProcessor(setID string) Processor {
	return Processor{
		setID:   setID,
		clients: map[schema.GroupVersionKind]Client{},
	}
}

func (t Processor) CodeVersion(version string) Processor {
	t.codeVersion = version
	return t
}

func (t Processor) Client(clients ...Client) Processor {
	// ensure cache is enabled
	for _, client := range clients {
		client.Generic()
		t.clients[client.ObjectClient().GroupVersionKind()] = client
	}
	return t
}

func (t Processor) Remove(owner runtime.Object) error {
	return t.NewDesiredSet(owner, nil).Apply()
}

func (t Processor) NewDesiredSet(owner runtime.Object, objs *ObjectSet) *DesiredSet {
	remove := false
	if objs == nil {
		remove = true
		objs = &ObjectSet{}
	}
	return &DesiredSet{
		remove:      remove,
		objs:        objs,
		setID:       t.setID,
		codeVersion: t.codeVersion,
		clients:     t.clients,
		owner:       owner,
	}
}

package client

import (
	"github.com/rancher/norman/types"
)

const (
	NodeType                      = "node"
	NodeFieldAddresses            = "addresses"
	NodeFieldAllocatable          = "allocatable"
	NodeFieldCapacity             = "capacity"
	NodeFieldConfigSource         = "configSource"
	NodeFieldCreated              = "created"
	NodeFieldDaemonEndpoints      = "daemonEndpoints"
	NodeFieldExternalID           = "externalID"
	NodeFieldImages               = "images"
	NodeFieldLabels               = "labels"
	NodeFieldName                 = "name"
	NodeFieldNodeInfo             = "nodeInfo"
	NodeFieldPhase                = "phase"
	NodeFieldPodCIDR              = "podCIDR"
	NodeFieldProviderID           = "providerID"
	NodeFieldRemoved              = "removed"
	NodeFieldState                = "state"
	NodeFieldTaints               = "taints"
	NodeFieldTransitioning        = "transitioning"
	NodeFieldTransitioningMessage = "transitioningMessage"
	NodeFieldUUID                 = "uuid"
	NodeFieldUnschedulable        = "unschedulable"
	NodeFieldVolumesAttached      = "volumesAttached"
	NodeFieldVolumesInUse         = "volumesInUse"
)

type Node struct {
	types.Resource
	Addresses            []NodeAddress        `json:"addresses,omitempty" yaml:"addresses,omitempty"`
	Allocatable          map[string]string    `json:"allocatable,omitempty" yaml:"allocatable,omitempty"`
	Capacity             map[string]string    `json:"capacity,omitempty" yaml:"capacity,omitempty"`
	ConfigSource         *NodeConfigSource    `json:"configSource,omitempty" yaml:"configSource,omitempty"`
	Created              string               `json:"created,omitempty" yaml:"created,omitempty"`
	DaemonEndpoints      *NodeDaemonEndpoints `json:"daemonEndpoints,omitempty" yaml:"daemonEndpoints,omitempty"`
	ExternalID           string               `json:"externalID,omitempty" yaml:"externalID,omitempty"`
	Images               []ContainerImage     `json:"images,omitempty" yaml:"images,omitempty"`
	Labels               map[string]string    `json:"labels,omitempty" yaml:"labels,omitempty"`
	Name                 string               `json:"name,omitempty" yaml:"name,omitempty"`
	NodeInfo             *NodeSystemInfo      `json:"nodeInfo,omitempty" yaml:"nodeInfo,omitempty"`
	Phase                string               `json:"phase,omitempty" yaml:"phase,omitempty"`
	PodCIDR              string               `json:"podCIDR,omitempty" yaml:"podCIDR,omitempty"`
	ProviderID           string               `json:"providerID,omitempty" yaml:"providerID,omitempty"`
	Removed              string               `json:"removed,omitempty" yaml:"removed,omitempty"`
	State                string               `json:"state,omitempty" yaml:"state,omitempty"`
	Taints               []Taint              `json:"taints,omitempty" yaml:"taints,omitempty"`
	Transitioning        string               `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`
	TransitioningMessage string               `json:"transitioningMessage,omitempty" yaml:"transitioningMessage,omitempty"`
	UUID                 string               `json:"uuid,omitempty" yaml:"uuid,omitempty"`
	Unschedulable        bool                 `json:"unschedulable,omitempty" yaml:"unschedulable,omitempty"`
	VolumesAttached      []AttachedVolume     `json:"volumesAttached,omitempty" yaml:"volumesAttached,omitempty"`
	VolumesInUse         []string             `json:"volumesInUse,omitempty" yaml:"volumesInUse,omitempty"`
}

type NodeCollection struct {
	types.Collection
	Data   []Node `json:"data,omitempty"`
	client *NodeClient
}

type NodeClient struct {
	apiClient *Client
}

type NodeOperations interface {
	List(opts *types.ListOpts) (*NodeCollection, error)
	Create(opts *Node) (*Node, error)
	Update(existing *Node, updates interface{}) (*Node, error)
	Replace(existing *Node) (*Node, error)
	ByID(id string) (*Node, error)
	Delete(container *Node) error
}

func newNodeClient(apiClient *Client) *NodeClient {
	return &NodeClient{
		apiClient: apiClient,
	}
}

func (c *NodeClient) Create(container *Node) (*Node, error) {
	resp := &Node{}
	err := c.apiClient.Ops.DoCreate(NodeType, container, resp)
	return resp, err
}

func (c *NodeClient) Update(existing *Node, updates interface{}) (*Node, error) {
	resp := &Node{}
	err := c.apiClient.Ops.DoUpdate(NodeType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *NodeClient) Replace(obj *Node) (*Node, error) {
	resp := &Node{}
	err := c.apiClient.Ops.DoReplace(NodeType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *NodeClient) List(opts *types.ListOpts) (*NodeCollection, error) {
	resp := &NodeCollection{}
	err := c.apiClient.Ops.DoList(NodeType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *NodeCollection) Next() (*NodeCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &NodeCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *NodeClient) ByID(id string) (*Node, error) {
	resp := &Node{}
	err := c.apiClient.Ops.DoByID(NodeType, id, resp)
	return resp, err
}

func (c *NodeClient) Delete(container *Node) error {
	return c.apiClient.Ops.DoResourceDelete(NodeType, &container.Resource)
}

package client

import (
	"github.com/rancher/norman/types"
)

const (
	SpaceType                      = "space"
	SpaceFieldCreated              = "created"
	SpaceFieldLabels               = "labels"
	SpaceFieldName                 = "name"
	SpaceFieldRemoved              = "removed"
	SpaceFieldState                = "state"
	SpaceFieldTransitioning        = "transitioning"
	SpaceFieldTransitioningMessage = "transitioningMessage"
	SpaceFieldUUID                 = "uuid"
)

type Space struct {
	types.Resource
	Created              string            `json:"created,omitempty" yaml:"created,omitempty"`
	Labels               map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Name                 string            `json:"name,omitempty" yaml:"name,omitempty"`
	Removed              string            `json:"removed,omitempty" yaml:"removed,omitempty"`
	State                string            `json:"state,omitempty" yaml:"state,omitempty"`
	Transitioning        string            `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`
	TransitioningMessage string            `json:"transitioningMessage,omitempty" yaml:"transitioningMessage,omitempty"`
	UUID                 string            `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type SpaceCollection struct {
	types.Collection
	Data   []Space `json:"data,omitempty"`
	client *SpaceClient
}

type SpaceClient struct {
	apiClient *Client
}

type SpaceOperations interface {
	List(opts *types.ListOpts) (*SpaceCollection, error)
	Create(opts *Space) (*Space, error)
	Update(existing *Space, updates interface{}) (*Space, error)
	Replace(existing *Space) (*Space, error)
	ByID(id string) (*Space, error)
	Delete(container *Space) error
}

func newSpaceClient(apiClient *Client) *SpaceClient {
	return &SpaceClient{
		apiClient: apiClient,
	}
}

func (c *SpaceClient) Create(container *Space) (*Space, error) {
	resp := &Space{}
	err := c.apiClient.Ops.DoCreate(SpaceType, container, resp)
	return resp, err
}

func (c *SpaceClient) Update(existing *Space, updates interface{}) (*Space, error) {
	resp := &Space{}
	err := c.apiClient.Ops.DoUpdate(SpaceType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *SpaceClient) Replace(obj *Space) (*Space, error) {
	resp := &Space{}
	err := c.apiClient.Ops.DoReplace(SpaceType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *SpaceClient) List(opts *types.ListOpts) (*SpaceCollection, error) {
	resp := &SpaceCollection{}
	err := c.apiClient.Ops.DoList(SpaceType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *SpaceCollection) Next() (*SpaceCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &SpaceCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *SpaceClient) ByID(id string) (*Space, error) {
	resp := &Space{}
	err := c.apiClient.Ops.DoByID(SpaceType, id, resp)
	return resp, err
}

func (c *SpaceClient) Delete(container *Space) error {
	return c.apiClient.Ops.DoResourceDelete(SpaceType, &container.Resource)
}

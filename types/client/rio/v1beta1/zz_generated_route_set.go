package client

import (
	"github.com/rancher/norman/types"
)

const (
	RouteSetType         = "routeSet"
	RouteSetFieldCreated = "created"
	RouteSetFieldLabels  = "labels"
	RouteSetFieldName    = "name"
	RouteSetFieldRemoved = "removed"
	RouteSetFieldRoutes  = "routes"
	RouteSetFieldSpaceID = "spaceId"
	RouteSetFieldStackID = "stackId"
	RouteSetFieldUUID    = "uuid"
)

type RouteSet struct {
	types.Resource
	Created string            `json:"created,omitempty" yaml:"created,omitempty"`
	Labels  map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Name    string            `json:"name,omitempty" yaml:"name,omitempty"`
	Removed string            `json:"removed,omitempty" yaml:"removed,omitempty"`
	Routes  []RouteSpec       `json:"routes,omitempty" yaml:"routes,omitempty"`
	SpaceID string            `json:"spaceId,omitempty" yaml:"spaceId,omitempty"`
	StackID string            `json:"stackId,omitempty" yaml:"stackId,omitempty"`
	UUID    string            `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type RouteSetCollection struct {
	types.Collection
	Data   []RouteSet `json:"data,omitempty"`
	client *RouteSetClient
}

type RouteSetClient struct {
	apiClient *Client
}

type RouteSetOperations interface {
	List(opts *types.ListOpts) (*RouteSetCollection, error)
	Create(opts *RouteSet) (*RouteSet, error)
	Update(existing *RouteSet, updates interface{}) (*RouteSet, error)
	Replace(existing *RouteSet) (*RouteSet, error)
	ByID(id string) (*RouteSet, error)
	Delete(container *RouteSet) error
}

func newRouteSetClient(apiClient *Client) *RouteSetClient {
	return &RouteSetClient{
		apiClient: apiClient,
	}
}

func (c *RouteSetClient) Create(container *RouteSet) (*RouteSet, error) {
	resp := &RouteSet{}
	err := c.apiClient.Ops.DoCreate(RouteSetType, container, resp)
	return resp, err
}

func (c *RouteSetClient) Update(existing *RouteSet, updates interface{}) (*RouteSet, error) {
	resp := &RouteSet{}
	err := c.apiClient.Ops.DoUpdate(RouteSetType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *RouteSetClient) Replace(obj *RouteSet) (*RouteSet, error) {
	resp := &RouteSet{}
	err := c.apiClient.Ops.DoReplace(RouteSetType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *RouteSetClient) List(opts *types.ListOpts) (*RouteSetCollection, error) {
	resp := &RouteSetCollection{}
	err := c.apiClient.Ops.DoList(RouteSetType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *RouteSetCollection) Next() (*RouteSetCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &RouteSetCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *RouteSetClient) ByID(id string) (*RouteSet, error) {
	resp := &RouteSet{}
	err := c.apiClient.Ops.DoByID(RouteSetType, id, resp)
	return resp, err
}

func (c *RouteSetClient) Delete(container *RouteSet) error {
	return c.apiClient.Ops.DoResourceDelete(RouteSetType, &container.Resource)
}

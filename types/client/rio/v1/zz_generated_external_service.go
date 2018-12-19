package client

import (
	"github.com/rancher/norman/types"
)

const (
	ExternalServiceType                      = "externalService"
	ExternalServiceFieldCreated              = "created"
	ExternalServiceFieldLabels               = "labels"
	ExternalServiceFieldName                 = "name"
	ExternalServiceFieldProjectID            = "projectId"
	ExternalServiceFieldRemoved              = "removed"
	ExternalServiceFieldStackID              = "stackId"
	ExternalServiceFieldState                = "state"
	ExternalServiceFieldTarget               = "target"
	ExternalServiceFieldTransitioning        = "transitioning"
	ExternalServiceFieldTransitioningMessage = "transitioningMessage"
	ExternalServiceFieldUUID                 = "uuid"
)

type ExternalService struct {
	types.Resource
	Created              string            `json:"created,omitempty" yaml:"created,omitempty"`
	Labels               map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Name                 string            `json:"name,omitempty" yaml:"name,omitempty"`
	ProjectID            string            `json:"projectId,omitempty" yaml:"projectId,omitempty"`
	Removed              string            `json:"removed,omitempty" yaml:"removed,omitempty"`
	StackID              string            `json:"stackId,omitempty" yaml:"stackId,omitempty"`
	State                string            `json:"state,omitempty" yaml:"state,omitempty"`
	Target               string            `json:"target,omitempty" yaml:"target,omitempty"`
	Transitioning        string            `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`
	TransitioningMessage string            `json:"transitioningMessage,omitempty" yaml:"transitioningMessage,omitempty"`
	UUID                 string            `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type ExternalServiceCollection struct {
	types.Collection
	Data   []ExternalService `json:"data,omitempty"`
	client *ExternalServiceClient
}

type ExternalServiceClient struct {
	apiClient *Client
}

type ExternalServiceOperations interface {
	List(opts *types.ListOpts) (*ExternalServiceCollection, error)
	Create(opts *ExternalService) (*ExternalService, error)
	Update(existing *ExternalService, updates interface{}) (*ExternalService, error)
	Replace(existing *ExternalService) (*ExternalService, error)
	ByID(id string) (*ExternalService, error)
	Delete(container *ExternalService) error
}

func newExternalServiceClient(apiClient *Client) *ExternalServiceClient {
	return &ExternalServiceClient{
		apiClient: apiClient,
	}
}

func (c *ExternalServiceClient) Create(container *ExternalService) (*ExternalService, error) {
	resp := &ExternalService{}
	err := c.apiClient.Ops.DoCreate(ExternalServiceType, container, resp)
	return resp, err
}

func (c *ExternalServiceClient) Update(existing *ExternalService, updates interface{}) (*ExternalService, error) {
	resp := &ExternalService{}
	err := c.apiClient.Ops.DoUpdate(ExternalServiceType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ExternalServiceClient) Replace(obj *ExternalService) (*ExternalService, error) {
	resp := &ExternalService{}
	err := c.apiClient.Ops.DoReplace(ExternalServiceType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *ExternalServiceClient) List(opts *types.ListOpts) (*ExternalServiceCollection, error) {
	resp := &ExternalServiceCollection{}
	err := c.apiClient.Ops.DoList(ExternalServiceType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ExternalServiceCollection) Next() (*ExternalServiceCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ExternalServiceCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ExternalServiceClient) ByID(id string) (*ExternalService, error) {
	resp := &ExternalService{}
	err := c.apiClient.Ops.DoByID(ExternalServiceType, id, resp)
	return resp, err
}

func (c *ExternalServiceClient) Delete(container *ExternalService) error {
	return c.apiClient.Ops.DoResourceDelete(ExternalServiceType, &container.Resource)
}

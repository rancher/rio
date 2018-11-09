package client

import (
	"github.com/rancher/norman/types"
)

const (
	PublicDomainType                = "publicDomain"
	PublicDomainFieldCreated        = "created"
	PublicDomainFieldDomainName     = "domainName"
	PublicDomainFieldLabels         = "labels"
	PublicDomainFieldName           = "name"
	PublicDomainFieldRemoved        = "removed"
	PublicDomainFieldRequestTLSCert = "requestTlsCert"
	PublicDomainFieldRouteSetName   = "routeSetName"
	PublicDomainFieldServiceName    = "serviceName"
	PublicDomainFieldSpaceName      = "spaceName"
	PublicDomainFieldStackName      = "stackName"
	PublicDomainFieldUUID           = "uuid"
)

type PublicDomain struct {
	types.Resource
	Created        string            `json:"created,omitempty" yaml:"created,omitempty"`
	DomainName     string            `json:"domainName,omitempty" yaml:"domainName,omitempty"`
	Labels         map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Name           string            `json:"name,omitempty" yaml:"name,omitempty"`
	Removed        string            `json:"removed,omitempty" yaml:"removed,omitempty"`
	RequestTLSCert bool              `json:"requestTlsCert,omitempty" yaml:"requestTlsCert,omitempty"`
	RouteSetName   string            `json:"routeSetName,omitempty" yaml:"routeSetName,omitempty"`
	ServiceName    string            `json:"serviceName,omitempty" yaml:"serviceName,omitempty"`
	SpaceName      string            `json:"spaceName,omitempty" yaml:"spaceName,omitempty"`
	StackName      string            `json:"stackName,omitempty" yaml:"stackName,omitempty"`
	UUID           string            `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type PublicDomainCollection struct {
	types.Collection
	Data   []PublicDomain `json:"data,omitempty"`
	client *PublicDomainClient
}

type PublicDomainClient struct {
	apiClient *Client
}

type PublicDomainOperations interface {
	List(opts *types.ListOpts) (*PublicDomainCollection, error)
	Create(opts *PublicDomain) (*PublicDomain, error)
	Update(existing *PublicDomain, updates interface{}) (*PublicDomain, error)
	Replace(existing *PublicDomain) (*PublicDomain, error)
	ByID(id string) (*PublicDomain, error)
	Delete(container *PublicDomain) error
}

func newPublicDomainClient(apiClient *Client) *PublicDomainClient {
	return &PublicDomainClient{
		apiClient: apiClient,
	}
}

func (c *PublicDomainClient) Create(container *PublicDomain) (*PublicDomain, error) {
	resp := &PublicDomain{}
	err := c.apiClient.Ops.DoCreate(PublicDomainType, container, resp)
	return resp, err
}

func (c *PublicDomainClient) Update(existing *PublicDomain, updates interface{}) (*PublicDomain, error) {
	resp := &PublicDomain{}
	err := c.apiClient.Ops.DoUpdate(PublicDomainType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *PublicDomainClient) Replace(obj *PublicDomain) (*PublicDomain, error) {
	resp := &PublicDomain{}
	err := c.apiClient.Ops.DoReplace(PublicDomainType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *PublicDomainClient) List(opts *types.ListOpts) (*PublicDomainCollection, error) {
	resp := &PublicDomainCollection{}
	err := c.apiClient.Ops.DoList(PublicDomainType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *PublicDomainCollection) Next() (*PublicDomainCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &PublicDomainCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *PublicDomainClient) ByID(id string) (*PublicDomain, error) {
	resp := &PublicDomain{}
	err := c.apiClient.Ops.DoByID(PublicDomainType, id, resp)
	return resp, err
}

func (c *PublicDomainClient) Delete(container *PublicDomain) error {
	return c.apiClient.Ops.DoResourceDelete(PublicDomainType, &container.Resource)
}

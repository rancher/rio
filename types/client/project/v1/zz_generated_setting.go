package client

import (
	"github.com/rancher/norman/types"
)

const (
	SettingType            = "setting"
	SettingFieldCreated    = "created"
	SettingFieldCustomized = "customized"
	SettingFieldDefault    = "default"
	SettingFieldLabels     = "labels"
	SettingFieldName       = "name"
	SettingFieldRemoved    = "removed"
	SettingFieldUUID       = "uuid"
	SettingFieldValue      = "value"
)

type Setting struct {
	types.Resource
	Created    string            `json:"created,omitempty" yaml:"created,omitempty"`
	Customized bool              `json:"customized,omitempty" yaml:"customized,omitempty"`
	Default    string            `json:"default,omitempty" yaml:"default,omitempty"`
	Labels     map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Name       string            `json:"name,omitempty" yaml:"name,omitempty"`
	Removed    string            `json:"removed,omitempty" yaml:"removed,omitempty"`
	UUID       string            `json:"uuid,omitempty" yaml:"uuid,omitempty"`
	Value      string            `json:"value,omitempty" yaml:"value,omitempty"`
}

type SettingCollection struct {
	types.Collection
	Data   []Setting `json:"data,omitempty"`
	client *SettingClient
}

type SettingClient struct {
	apiClient *Client
}

type SettingOperations interface {
	List(opts *types.ListOpts) (*SettingCollection, error)
	Create(opts *Setting) (*Setting, error)
	Update(existing *Setting, updates interface{}) (*Setting, error)
	Replace(existing *Setting) (*Setting, error)
	ByID(id string) (*Setting, error)
	Delete(container *Setting) error
}

func newSettingClient(apiClient *Client) *SettingClient {
	return &SettingClient{
		apiClient: apiClient,
	}
}

func (c *SettingClient) Create(container *Setting) (*Setting, error) {
	resp := &Setting{}
	err := c.apiClient.Ops.DoCreate(SettingType, container, resp)
	return resp, err
}

func (c *SettingClient) Update(existing *Setting, updates interface{}) (*Setting, error) {
	resp := &Setting{}
	err := c.apiClient.Ops.DoUpdate(SettingType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *SettingClient) Replace(obj *Setting) (*Setting, error) {
	resp := &Setting{}
	err := c.apiClient.Ops.DoReplace(SettingType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *SettingClient) List(opts *types.ListOpts) (*SettingCollection, error) {
	resp := &SettingCollection{}
	err := c.apiClient.Ops.DoList(SettingType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *SettingCollection) Next() (*SettingCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &SettingCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *SettingClient) ByID(id string) (*Setting, error) {
	resp := &Setting{}
	err := c.apiClient.Ops.DoByID(SettingType, id, resp)
	return resp, err
}

func (c *SettingClient) Delete(container *Setting) error {
	return c.apiClient.Ops.DoResourceDelete(SettingType, &container.Resource)
}

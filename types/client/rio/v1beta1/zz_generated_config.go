package client

import (
	"github.com/rancher/norman/types"
)

const (
	ConfigType             = "config"
	ConfigFieldContent     = "content"
	ConfigFieldCreated     = "created"
	ConfigFieldDescription = "description"
	ConfigFieldEncoded     = "encoded"
	ConfigFieldLabels      = "labels"
	ConfigFieldName        = "name"
	ConfigFieldRemoved     = "removed"
	ConfigFieldSpaceID     = "spaceId"
	ConfigFieldStackID     = "stackId"
	ConfigFieldUUID        = "uuid"
)

type Config struct {
	types.Resource
	Content     string            `json:"content,omitempty" yaml:"content,omitempty"`
	Created     string            `json:"created,omitempty" yaml:"created,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Encoded     bool              `json:"encoded,omitempty" yaml:"encoded,omitempty"`
	Labels      map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Name        string            `json:"name,omitempty" yaml:"name,omitempty"`
	Removed     string            `json:"removed,omitempty" yaml:"removed,omitempty"`
	SpaceID     string            `json:"spaceId,omitempty" yaml:"spaceId,omitempty"`
	StackID     string            `json:"stackId,omitempty" yaml:"stackId,omitempty"`
	UUID        string            `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type ConfigCollection struct {
	types.Collection
	Data   []Config `json:"data,omitempty"`
	client *ConfigClient
}

type ConfigClient struct {
	apiClient *Client
}

type ConfigOperations interface {
	List(opts *types.ListOpts) (*ConfigCollection, error)
	Create(opts *Config) (*Config, error)
	Update(existing *Config, updates interface{}) (*Config, error)
	Replace(existing *Config) (*Config, error)
	ByID(id string) (*Config, error)
	Delete(container *Config) error
}

func newConfigClient(apiClient *Client) *ConfigClient {
	return &ConfigClient{
		apiClient: apiClient,
	}
}

func (c *ConfigClient) Create(container *Config) (*Config, error) {
	resp := &Config{}
	err := c.apiClient.Ops.DoCreate(ConfigType, container, resp)
	return resp, err
}

func (c *ConfigClient) Update(existing *Config, updates interface{}) (*Config, error) {
	resp := &Config{}
	err := c.apiClient.Ops.DoUpdate(ConfigType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ConfigClient) Replace(obj *Config) (*Config, error) {
	resp := &Config{}
	err := c.apiClient.Ops.DoReplace(ConfigType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *ConfigClient) List(opts *types.ListOpts) (*ConfigCollection, error) {
	resp := &ConfigCollection{}
	err := c.apiClient.Ops.DoList(ConfigType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ConfigCollection) Next() (*ConfigCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ConfigCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ConfigClient) ByID(id string) (*Config, error) {
	resp := &Config{}
	err := c.apiClient.Ops.DoByID(ConfigType, id, resp)
	return resp, err
}

func (c *ConfigClient) Delete(container *Config) error {
	return c.apiClient.Ops.DoResourceDelete(ConfigType, &container.Resource)
}

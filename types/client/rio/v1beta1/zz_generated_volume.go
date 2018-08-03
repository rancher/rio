package client

import (
	"github.com/rancher/norman/types"
)

const (
	VolumeType                      = "volume"
	VolumeFieldAccessMode           = "accessMode"
	VolumeFieldAccessModes          = "accessModes"
	VolumeFieldCapacity             = "capacity"
	VolumeFieldCreated              = "created"
	VolumeFieldDescription          = "description"
	VolumeFieldDriver               = "driver"
	VolumeFieldLabels               = "labels"
	VolumeFieldName                 = "name"
	VolumeFieldRemoved              = "removed"
	VolumeFieldSizeInGB             = "sizeInGb"
	VolumeFieldSpaceID              = "spaceId"
	VolumeFieldStackID              = "stackId"
	VolumeFieldState                = "state"
	VolumeFieldTemplate             = "template"
	VolumeFieldTransitioning        = "transitioning"
	VolumeFieldTransitioningMessage = "transitioningMessage"
	VolumeFieldUUID                 = "uuid"
)

type Volume struct {
	types.Resource
	AccessMode           string            `json:"accessMode,omitempty" yaml:"accessMode,omitempty"`
	AccessModes          []string          `json:"accessModes,omitempty" yaml:"accessModes,omitempty"`
	Capacity             map[string]string `json:"capacity,omitempty" yaml:"capacity,omitempty"`
	Created              string            `json:"created,omitempty" yaml:"created,omitempty"`
	Description          string            `json:"description,omitempty" yaml:"description,omitempty"`
	Driver               string            `json:"driver,omitempty" yaml:"driver,omitempty"`
	Labels               map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Name                 string            `json:"name,omitempty" yaml:"name,omitempty"`
	Removed              string            `json:"removed,omitempty" yaml:"removed,omitempty"`
	SizeInGB             int64             `json:"sizeInGb,omitempty" yaml:"sizeInGb,omitempty"`
	SpaceID              string            `json:"spaceId,omitempty" yaml:"spaceId,omitempty"`
	StackID              string            `json:"stackId,omitempty" yaml:"stackId,omitempty"`
	State                string            `json:"state,omitempty" yaml:"state,omitempty"`
	Template             bool              `json:"template,omitempty" yaml:"template,omitempty"`
	Transitioning        string            `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`
	TransitioningMessage string            `json:"transitioningMessage,omitempty" yaml:"transitioningMessage,omitempty"`
	UUID                 string            `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type VolumeCollection struct {
	types.Collection
	Data   []Volume `json:"data,omitempty"`
	client *VolumeClient
}

type VolumeClient struct {
	apiClient *Client
}

type VolumeOperations interface {
	List(opts *types.ListOpts) (*VolumeCollection, error)
	Create(opts *Volume) (*Volume, error)
	Update(existing *Volume, updates interface{}) (*Volume, error)
	Replace(existing *Volume) (*Volume, error)
	ByID(id string) (*Volume, error)
	Delete(container *Volume) error
}

func newVolumeClient(apiClient *Client) *VolumeClient {
	return &VolumeClient{
		apiClient: apiClient,
	}
}

func (c *VolumeClient) Create(container *Volume) (*Volume, error) {
	resp := &Volume{}
	err := c.apiClient.Ops.DoCreate(VolumeType, container, resp)
	return resp, err
}

func (c *VolumeClient) Update(existing *Volume, updates interface{}) (*Volume, error) {
	resp := &Volume{}
	err := c.apiClient.Ops.DoUpdate(VolumeType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *VolumeClient) Replace(obj *Volume) (*Volume, error) {
	resp := &Volume{}
	err := c.apiClient.Ops.DoReplace(VolumeType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *VolumeClient) List(opts *types.ListOpts) (*VolumeCollection, error) {
	resp := &VolumeCollection{}
	err := c.apiClient.Ops.DoList(VolumeType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *VolumeCollection) Next() (*VolumeCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &VolumeCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *VolumeClient) ByID(id string) (*Volume, error) {
	resp := &Volume{}
	err := c.apiClient.Ops.DoByID(VolumeType, id, resp)
	return resp, err
}

func (c *VolumeClient) Delete(container *Volume) error {
	return c.apiClient.Ops.DoResourceDelete(VolumeType, &container.Resource)
}

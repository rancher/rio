package client

import (
	"github.com/rancher/norman/types"
)

const (
	FeatureType                      = "feature"
	FeatureFieldAnswers              = "answers"
	FeatureFieldCreated              = "created"
	FeatureFieldDescription          = "description"
	FeatureFieldEnabled              = "enable"
	FeatureFieldLabels               = "labels"
	FeatureFieldName                 = "name"
	FeatureFieldNamespace            = "namespace"
	FeatureFieldQuestions            = "questions"
	FeatureFieldRemoved              = "removed"
	FeatureFieldState                = "state"
	FeatureFieldTransitioning        = "transitioning"
	FeatureFieldTransitioningMessage = "transitioningMessage"
	FeatureFieldUUID                 = "uuid"
)

type Feature struct {
	types.Resource
	Answers              map[string]string `json:"answers,omitempty" yaml:"answers,omitempty"`
	Created              string            `json:"created,omitempty" yaml:"created,omitempty"`
	Description          string            `json:"description,omitempty" yaml:"description,omitempty"`
	Enabled              bool              `json:"enable,omitempty" yaml:"enable,omitempty"`
	Labels               map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Name                 string            `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace            string            `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Questions            []Question        `json:"questions,omitempty" yaml:"questions,omitempty"`
	Removed              string            `json:"removed,omitempty" yaml:"removed,omitempty"`
	State                string            `json:"state,omitempty" yaml:"state,omitempty"`
	Transitioning        string            `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`
	TransitioningMessage string            `json:"transitioningMessage,omitempty" yaml:"transitioningMessage,omitempty"`
	UUID                 string            `json:"uuid,omitempty" yaml:"uuid,omitempty"`
}

type FeatureCollection struct {
	types.Collection
	Data   []Feature `json:"data,omitempty"`
	client *FeatureClient
}

type FeatureClient struct {
	apiClient *Client
}

type FeatureOperations interface {
	List(opts *types.ListOpts) (*FeatureCollection, error)
	Create(opts *Feature) (*Feature, error)
	Update(existing *Feature, updates interface{}) (*Feature, error)
	Replace(existing *Feature) (*Feature, error)
	ByID(id string) (*Feature, error)
	Delete(container *Feature) error
}

func newFeatureClient(apiClient *Client) *FeatureClient {
	return &FeatureClient{
		apiClient: apiClient,
	}
}

func (c *FeatureClient) Create(container *Feature) (*Feature, error) {
	resp := &Feature{}
	err := c.apiClient.Ops.DoCreate(FeatureType, container, resp)
	return resp, err
}

func (c *FeatureClient) Update(existing *Feature, updates interface{}) (*Feature, error) {
	resp := &Feature{}
	err := c.apiClient.Ops.DoUpdate(FeatureType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *FeatureClient) Replace(obj *Feature) (*Feature, error) {
	resp := &Feature{}
	err := c.apiClient.Ops.DoReplace(FeatureType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *FeatureClient) List(opts *types.ListOpts) (*FeatureCollection, error) {
	resp := &FeatureCollection{}
	err := c.apiClient.Ops.DoList(FeatureType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *FeatureCollection) Next() (*FeatureCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &FeatureCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *FeatureClient) ByID(id string) (*Feature, error) {
	resp := &Feature{}
	err := c.apiClient.Ops.DoByID(FeatureType, id, resp)
	return resp, err
}

func (c *FeatureClient) Delete(container *Feature) error {
	return c.apiClient.Ops.DoResourceDelete(FeatureType, &container.Resource)
}

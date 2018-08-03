package client

import (
	"github.com/rancher/norman/types"
)

const (
	PodType                               = "pod"
	PodFieldActiveDeadlineSeconds         = "activeDeadlineSeconds"
	PodFieldAffinity                      = "affinity"
	PodFieldAutomountServiceAccountToken  = "automountServiceAccountToken"
	PodFieldContainerStatuses             = "containerStatuses"
	PodFieldContainers                    = "containers"
	PodFieldCreated                       = "created"
	PodFieldDNSConfig                     = "dnsConfig"
	PodFieldDNSPolicy                     = "dnsPolicy"
	PodFieldDeprecatedServiceAccount      = "serviceAccount"
	PodFieldHostAliases                   = "hostAliases"
	PodFieldHostIP                        = "hostIP"
	PodFieldHostIPC                       = "hostIPC"
	PodFieldHostNetwork                   = "hostNetwork"
	PodFieldHostPID                       = "hostPID"
	PodFieldHostname                      = "hostname"
	PodFieldImagePullSecrets              = "imagePullSecrets"
	PodFieldInitContainerStatuses         = "initContainerStatuses"
	PodFieldInitContainers                = "initContainers"
	PodFieldLabels                        = "labels"
	PodFieldMessage                       = "message"
	PodFieldName                          = "name"
	PodFieldNamespace                     = "namespace"
	PodFieldNodeName                      = "nodeName"
	PodFieldNodeSelector                  = "nodeSelector"
	PodFieldNominatedNodeName             = "nominatedNodeName"
	PodFieldPhase                         = "phase"
	PodFieldPodIP                         = "podIP"
	PodFieldPriority                      = "priority"
	PodFieldPriorityClassName             = "priorityClassName"
	PodFieldQOSClass                      = "qosClass"
	PodFieldReason                        = "reason"
	PodFieldRemoved                       = "removed"
	PodFieldRestartPolicy                 = "restartPolicy"
	PodFieldSchedulerName                 = "schedulerName"
	PodFieldSecurityContext               = "securityContext"
	PodFieldServiceAccountName            = "serviceAccountName"
	PodFieldShareProcessNamespace         = "shareProcessNamespace"
	PodFieldStartTime                     = "startTime"
	PodFieldState                         = "state"
	PodFieldSubdomain                     = "subdomain"
	PodFieldTerminationGracePeriodSeconds = "terminationGracePeriodSeconds"
	PodFieldTolerations                   = "tolerations"
	PodFieldTransitioning                 = "transitioning"
	PodFieldTransitioningMessage          = "transitioningMessage"
	PodFieldUUID                          = "uuid"
	PodFieldVolumes                       = "volumes"
)

type Pod struct {
	types.Resource
	ActiveDeadlineSeconds         *int64                 `json:"activeDeadlineSeconds,omitempty" yaml:"activeDeadlineSeconds,omitempty"`
	Affinity                      *Affinity              `json:"affinity,omitempty" yaml:"affinity,omitempty"`
	AutomountServiceAccountToken  *bool                  `json:"automountServiceAccountToken,omitempty" yaml:"automountServiceAccountToken,omitempty"`
	ContainerStatuses             []ContainerStatus      `json:"containerStatuses,omitempty" yaml:"containerStatuses,omitempty"`
	Containers                    []Container            `json:"containers,omitempty" yaml:"containers,omitempty"`
	Created                       string                 `json:"created,omitempty" yaml:"created,omitempty"`
	DNSConfig                     *PodDNSConfig          `json:"dnsConfig,omitempty" yaml:"dnsConfig,omitempty"`
	DNSPolicy                     string                 `json:"dnsPolicy,omitempty" yaml:"dnsPolicy,omitempty"`
	DeprecatedServiceAccount      string                 `json:"serviceAccount,omitempty" yaml:"serviceAccount,omitempty"`
	HostAliases                   []HostAlias            `json:"hostAliases,omitempty" yaml:"hostAliases,omitempty"`
	HostIP                        string                 `json:"hostIP,omitempty" yaml:"hostIP,omitempty"`
	HostIPC                       bool                   `json:"hostIPC,omitempty" yaml:"hostIPC,omitempty"`
	HostNetwork                   bool                   `json:"hostNetwork,omitempty" yaml:"hostNetwork,omitempty"`
	HostPID                       bool                   `json:"hostPID,omitempty" yaml:"hostPID,omitempty"`
	Hostname                      string                 `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	ImagePullSecrets              []LocalObjectReference `json:"imagePullSecrets,omitempty" yaml:"imagePullSecrets,omitempty"`
	InitContainerStatuses         []ContainerStatus      `json:"initContainerStatuses,omitempty" yaml:"initContainerStatuses,omitempty"`
	InitContainers                []Container            `json:"initContainers,omitempty" yaml:"initContainers,omitempty"`
	Labels                        map[string]string      `json:"labels,omitempty" yaml:"labels,omitempty"`
	Message                       string                 `json:"message,omitempty" yaml:"message,omitempty"`
	Name                          string                 `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace                     string                 `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	NodeName                      string                 `json:"nodeName,omitempty" yaml:"nodeName,omitempty"`
	NodeSelector                  map[string]string      `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	NominatedNodeName             string                 `json:"nominatedNodeName,omitempty" yaml:"nominatedNodeName,omitempty"`
	Phase                         string                 `json:"phase,omitempty" yaml:"phase,omitempty"`
	PodIP                         string                 `json:"podIP,omitempty" yaml:"podIP,omitempty"`
	Priority                      *int64                 `json:"priority,omitempty" yaml:"priority,omitempty"`
	PriorityClassName             string                 `json:"priorityClassName,omitempty" yaml:"priorityClassName,omitempty"`
	QOSClass                      string                 `json:"qosClass,omitempty" yaml:"qosClass,omitempty"`
	Reason                        string                 `json:"reason,omitempty" yaml:"reason,omitempty"`
	Removed                       string                 `json:"removed,omitempty" yaml:"removed,omitempty"`
	RestartPolicy                 string                 `json:"restartPolicy,omitempty" yaml:"restartPolicy,omitempty"`
	SchedulerName                 string                 `json:"schedulerName,omitempty" yaml:"schedulerName,omitempty"`
	SecurityContext               *PodSecurityContext    `json:"securityContext,omitempty" yaml:"securityContext,omitempty"`
	ServiceAccountName            string                 `json:"serviceAccountName,omitempty" yaml:"serviceAccountName,omitempty"`
	ShareProcessNamespace         *bool                  `json:"shareProcessNamespace,omitempty" yaml:"shareProcessNamespace,omitempty"`
	StartTime                     string                 `json:"startTime,omitempty" yaml:"startTime,omitempty"`
	State                         string                 `json:"state,omitempty" yaml:"state,omitempty"`
	Subdomain                     string                 `json:"subdomain,omitempty" yaml:"subdomain,omitempty"`
	TerminationGracePeriodSeconds *int64                 `json:"terminationGracePeriodSeconds,omitempty" yaml:"terminationGracePeriodSeconds,omitempty"`
	Tolerations                   []Toleration           `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
	Transitioning                 string                 `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`
	TransitioningMessage          string                 `json:"transitioningMessage,omitempty" yaml:"transitioningMessage,omitempty"`
	UUID                          string                 `json:"uuid,omitempty" yaml:"uuid,omitempty"`
	Volumes                       []Volume               `json:"volumes,omitempty" yaml:"volumes,omitempty"`
}

type PodCollection struct {
	types.Collection
	Data   []Pod `json:"data,omitempty"`
	client *PodClient
}

type PodClient struct {
	apiClient *Client
}

type PodOperations interface {
	List(opts *types.ListOpts) (*PodCollection, error)
	Create(opts *Pod) (*Pod, error)
	Update(existing *Pod, updates interface{}) (*Pod, error)
	Replace(existing *Pod) (*Pod, error)
	ByID(id string) (*Pod, error)
	Delete(container *Pod) error
}

func newPodClient(apiClient *Client) *PodClient {
	return &PodClient{
		apiClient: apiClient,
	}
}

func (c *PodClient) Create(container *Pod) (*Pod, error) {
	resp := &Pod{}
	err := c.apiClient.Ops.DoCreate(PodType, container, resp)
	return resp, err
}

func (c *PodClient) Update(existing *Pod, updates interface{}) (*Pod, error) {
	resp := &Pod{}
	err := c.apiClient.Ops.DoUpdate(PodType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *PodClient) Replace(obj *Pod) (*Pod, error) {
	resp := &Pod{}
	err := c.apiClient.Ops.DoReplace(PodType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *PodClient) List(opts *types.ListOpts) (*PodCollection, error) {
	resp := &PodCollection{}
	err := c.apiClient.Ops.DoList(PodType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *PodCollection) Next() (*PodCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &PodCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *PodClient) ByID(id string) (*Pod, error) {
	resp := &Pod{}
	err := c.apiClient.Ops.DoByID(PodType, id, resp)
	return resp, err
}

func (c *PodClient) Delete(container *Pod) error {
	return c.apiClient.Ops.DoResourceDelete(PodType, &container.Resource)
}

package client

import (
	"github.com/rancher/norman/types"
)

const (
	ServiceType                        = "service"
	ServiceFieldBatchSize              = "batchSize"
	ServiceFieldCPUs                   = "nanoCpus"
	ServiceFieldCapAdd                 = "capAdd"
	ServiceFieldCapDrop                = "capDrop"
	ServiceFieldCommand                = "command"
	ServiceFieldConfigs                = "configs"
	ServiceFieldCreated                = "created"
	ServiceFieldDNS                    = "dns"
	ServiceFieldDNSOptions             = "dnsOptions"
	ServiceFieldDNSSearch              = "dnsSearch"
	ServiceFieldDefaultVolumeDriver    = "defaultVolumeDriver"
	ServiceFieldDeploymentStrategy     = "deploymentStrategy"
	ServiceFieldDevices                = "devices"
	ServiceFieldEntrypoint             = "entrypoint"
	ServiceFieldEnvironment            = "environment"
	ServiceFieldExposedPorts           = "expose"
	ServiceFieldExtraHosts             = "extraHosts"
	ServiceFieldGlobal                 = "global"
	ServiceFieldGlobalPermissions      = "globalPermissions"
	ServiceFieldHealthcheck            = "healthcheck"
	ServiceFieldHostname               = "hostname"
	ServiceFieldImage                  = "image"
	ServiceFieldImagePullPolicy        = "imagePullPolicy"
	ServiceFieldInit                   = "init"
	ServiceFieldIpcMode                = "ipc"
	ServiceFieldLabels                 = "labels"
	ServiceFieldMemoryLimitBytes       = "memoryLimitBytes"
	ServiceFieldMemoryReservationBytes = "memoryReservationBytes"
	ServiceFieldMetadata               = "metadata"
	ServiceFieldName                   = "name"
	ServiceFieldNetworkMode            = "net"
	ServiceFieldOpenStdin              = "stdinOpen"
	ServiceFieldPermissions            = "permissions"
	ServiceFieldPidMode                = "pid"
	ServiceFieldPortBindings           = "ports"
	ServiceFieldPrivileged             = "privileged"
	ServiceFieldReadonlyRootfs         = "readOnly"
	ServiceFieldRemoved                = "removed"
	ServiceFieldRestartPolicy          = "restart"
	ServiceFieldRevisions              = "revisions"
	ServiceFieldScale                  = "scale"
	ServiceFieldScaleStatus            = "scaleStatus"
	ServiceFieldScheduling             = "scheduling"
	ServiceFieldSecrets                = "secrets"
	ServiceFieldSidekicks              = "sidekicks"
	ServiceFieldSpaceID                = "spaceId"
	ServiceFieldStackID                = "stackId"
	ServiceFieldState                  = "state"
	ServiceFieldStopGracePeriodSeconds = "stopGracePeriod"
	ServiceFieldTmpfs                  = "tmpfs"
	ServiceFieldTransitioning          = "transitioning"
	ServiceFieldTransitioningMessage   = "transitioningMessage"
	ServiceFieldTty                    = "tty"
	ServiceFieldUUID                   = "uuid"
	ServiceFieldUpdateOrder            = "updateOrder"
	ServiceFieldUpdateStrategy         = "updateStrategy"
	ServiceFieldUser                   = "user"
	ServiceFieldVolumes                = "volumes"
	ServiceFieldVolumesFrom            = "volumesFrom"
	ServiceFieldWorkingDir             = "workingDir"
)

type Service struct {
	types.Resource
	BatchSize              int64                      `json:"batchSize,omitempty" yaml:"batchSize,omitempty"`
	CPUs                   string                     `json:"nanoCpus,omitempty" yaml:"nanoCpus,omitempty"`
	CapAdd                 []string                   `json:"capAdd,omitempty" yaml:"capAdd,omitempty"`
	CapDrop                []string                   `json:"capDrop,omitempty" yaml:"capDrop,omitempty"`
	Command                []string                   `json:"command,omitempty" yaml:"command,omitempty"`
	Configs                []ConfigMapping            `json:"configs,omitempty" yaml:"configs,omitempty"`
	Created                string                     `json:"created,omitempty" yaml:"created,omitempty"`
	DNS                    []string                   `json:"dns,omitempty" yaml:"dns,omitempty"`
	DNSOptions             []string                   `json:"dnsOptions,omitempty" yaml:"dnsOptions,omitempty"`
	DNSSearch              []string                   `json:"dnsSearch,omitempty" yaml:"dnsSearch,omitempty"`
	DefaultVolumeDriver    string                     `json:"defaultVolumeDriver,omitempty" yaml:"defaultVolumeDriver,omitempty"`
	DeploymentStrategy     string                     `json:"deploymentStrategy,omitempty" yaml:"deploymentStrategy,omitempty"`
	Devices                []DeviceMapping            `json:"devices,omitempty" yaml:"devices,omitempty"`
	Entrypoint             []string                   `json:"entrypoint,omitempty" yaml:"entrypoint,omitempty"`
	Environment            []string                   `json:"environment,omitempty" yaml:"environment,omitempty"`
	ExposedPorts           []ExposedPort              `json:"expose,omitempty" yaml:"expose,omitempty"`
	ExtraHosts             []string                   `json:"extraHosts,omitempty" yaml:"extraHosts,omitempty"`
	Global                 bool                       `json:"global,omitempty" yaml:"global,omitempty"`
	GlobalPermissions      []Permission               `json:"globalPermissions,omitempty" yaml:"globalPermissions,omitempty"`
	Healthcheck            *HealthConfig              `json:"healthcheck,omitempty" yaml:"healthcheck,omitempty"`
	Hostname               string                     `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	Image                  string                     `json:"image,omitempty" yaml:"image,omitempty"`
	ImagePullPolicy        string                     `json:"imagePullPolicy,omitempty" yaml:"imagePullPolicy,omitempty"`
	Init                   bool                       `json:"init,omitempty" yaml:"init,omitempty"`
	IpcMode                string                     `json:"ipc,omitempty" yaml:"ipc,omitempty"`
	Labels                 map[string]string          `json:"labels,omitempty" yaml:"labels,omitempty"`
	MemoryLimitBytes       int64                      `json:"memoryLimitBytes,omitempty" yaml:"memoryLimitBytes,omitempty"`
	MemoryReservationBytes int64                      `json:"memoryReservationBytes,omitempty" yaml:"memoryReservationBytes,omitempty"`
	Metadata               map[string]interface{}     `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Name                   string                     `json:"name,omitempty" yaml:"name,omitempty"`
	NetworkMode            string                     `json:"net,omitempty" yaml:"net,omitempty"`
	OpenStdin              bool                       `json:"stdinOpen,omitempty" yaml:"stdinOpen,omitempty"`
	Permissions            []Permission               `json:"permissions,omitempty" yaml:"permissions,omitempty"`
	PidMode                string                     `json:"pid,omitempty" yaml:"pid,omitempty"`
	PortBindings           []PortBinding              `json:"ports,omitempty" yaml:"ports,omitempty"`
	Privileged             bool                       `json:"privileged,omitempty" yaml:"privileged,omitempty"`
	ReadonlyRootfs         bool                       `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	Removed                string                     `json:"removed,omitempty" yaml:"removed,omitempty"`
	RestartPolicy          string                     `json:"restart,omitempty" yaml:"restart,omitempty"`
	Revisions              map[string]ServiceRevision `json:"revisions,omitempty" yaml:"revisions,omitempty"`
	Scale                  int64                      `json:"scale,omitempty" yaml:"scale,omitempty"`
	ScaleStatus            *ScaleStatus               `json:"scaleStatus,omitempty" yaml:"scaleStatus,omitempty"`
	Scheduling             *Scheduling                `json:"scheduling,omitempty" yaml:"scheduling,omitempty"`
	Secrets                []SecretMapping            `json:"secrets,omitempty" yaml:"secrets,omitempty"`
	Sidekicks              map[string]SidekickConfig  `json:"sidekicks,omitempty" yaml:"sidekicks,omitempty"`
	SpaceID                string                     `json:"spaceId,omitempty" yaml:"spaceId,omitempty"`
	StackID                string                     `json:"stackId,omitempty" yaml:"stackId,omitempty"`
	State                  string                     `json:"state,omitempty" yaml:"state,omitempty"`
	StopGracePeriodSeconds *int64                     `json:"stopGracePeriod,omitempty" yaml:"stopGracePeriod,omitempty"`
	Tmpfs                  []Tmpfs                    `json:"tmpfs,omitempty" yaml:"tmpfs,omitempty"`
	Transitioning          string                     `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`
	TransitioningMessage   string                     `json:"transitioningMessage,omitempty" yaml:"transitioningMessage,omitempty"`
	Tty                    bool                       `json:"tty,omitempty" yaml:"tty,omitempty"`
	UUID                   string                     `json:"uuid,omitempty" yaml:"uuid,omitempty"`
	UpdateOrder            string                     `json:"updateOrder,omitempty" yaml:"updateOrder,omitempty"`
	UpdateStrategy         string                     `json:"updateStrategy,omitempty" yaml:"updateStrategy,omitempty"`
	User                   string                     `json:"user,omitempty" yaml:"user,omitempty"`
	Volumes                []Mount                    `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	VolumesFrom            []string                   `json:"volumesFrom,omitempty" yaml:"volumesFrom,omitempty"`
	WorkingDir             string                     `json:"workingDir,omitempty" yaml:"workingDir,omitempty"`
}

type ServiceCollection struct {
	types.Collection
	Data   []Service `json:"data,omitempty"`
	client *ServiceClient
}

type ServiceClient struct {
	apiClient *Client
}

type ServiceOperations interface {
	List(opts *types.ListOpts) (*ServiceCollection, error)
	Create(opts *Service) (*Service, error)
	Update(existing *Service, updates interface{}) (*Service, error)
	Replace(existing *Service) (*Service, error)
	ByID(id string) (*Service, error)
	Delete(container *Service) error
}

func newServiceClient(apiClient *Client) *ServiceClient {
	return &ServiceClient{
		apiClient: apiClient,
	}
}

func (c *ServiceClient) Create(container *Service) (*Service, error) {
	resp := &Service{}
	err := c.apiClient.Ops.DoCreate(ServiceType, container, resp)
	return resp, err
}

func (c *ServiceClient) Update(existing *Service, updates interface{}) (*Service, error) {
	resp := &Service{}
	err := c.apiClient.Ops.DoUpdate(ServiceType, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ServiceClient) Replace(obj *Service) (*Service, error) {
	resp := &Service{}
	err := c.apiClient.Ops.DoReplace(ServiceType, &obj.Resource, obj, resp)
	return resp, err
}

func (c *ServiceClient) List(opts *types.ListOpts) (*ServiceCollection, error) {
	resp := &ServiceCollection{}
	err := c.apiClient.Ops.DoList(ServiceType, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ServiceCollection) Next() (*ServiceCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ServiceCollection{}
		err := cc.client.apiClient.Ops.DoNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ServiceClient) ByID(id string) (*Service, error) {
	resp := &Service{}
	err := c.apiClient.Ops.DoByID(ServiceType, id, resp)
	return resp, err
}

func (c *ServiceClient) Delete(container *Service) error {
	return c.apiClient.Ops.DoResourceDelete(ServiceType, &container.Resource)
}

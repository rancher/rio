package create

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/stack"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/kvfile"
	"github.com/rancher/rio/cli/pkg/waiter"
	"github.com/rancher/rio/types/client/rio/v1beta1"
)

type Create struct {
	AddHost            []string          `desc:"Add a custom host-to-IP mapping (host:ip)"`
	CapAdd             []string          `desc:"Add Linux capabilities"`
	CapDrop            []string          `desc:"Drop Linux capabilities"`
	Config             []string          `desc:"Configs to expose to the service (format: name:target)"`
	Cpus               string            `desc:"Number of CPUs"`
	DeploymentStrategy string            `desc:"Approach to creating containers (parallel|ordered)" default:"parallel"`
	Detach             bool              `desc:"Do not attach after when -it is specified"`
	Device             []string          `desc:"Add a host device to the container"`
	DnsOption          []string          `desc:"Set DNS options"`
	DnsSearch          []string          `desc:"Set custom DNS search domains"`
	Dns                []string          `desc:"Set custom DNS servers"`
	E_Env              []string          `desc:"Set environment variables"`
	Entrypoint         []string          `desc:"Overwrite the default ENTRYPOINT of the image"`
	EnvFile            []string          `desc:"Read in a file of environment variables"`
	Expose             []string          `desc:"Expose a container's port(s) internally"`
	GlobalPermission   []string          `desc:"Permissions to grant to container's service account for all stacks"`
	GroupAdd           []string          `desc:"Add additional groups to join"`
	HealthCmd          string            `desc:"Command to run to check health"`
	HealthInterval     string            `desc:"Time between running the check (ms|s|m|h)" default:"0s"`
	HealthRetries      int               `desc:"Consecutive successes needed to report healthy"`
	HealthStartPeriod  string            `desc:"Start period for the container to initialize before starting healthchecks (ms|s|m|h)" default:"0s"`
	HealthTimeout      string            `desc:"Maximum time to allow one check to run (ms|s|m|h)" default:"0s"`
	HealthURL          string            `desc:"URL to hit to check health (example: http://localhost:8080/ping)"`
	Hostname           string            `desc:"Container host name"`
	I_Interactive      bool              `desc:"Keep STDIN open even if not attached"`
	ImagePullPolicy    string            `desc:"Behavior determining when to pull the image (never|always|not-present)" default:"not-present"`
	Init               bool              `desc:"Run an init inside the container that forwards signals and reaps processes"`
	Ipc                string            `desc:"IPC mode to use"`
	LabelFile          []string          `desc:"Read in a line delimited file of labels"`
	L_Label            map[string]string `desc:"Set meta data on a container"`
	MemoryLimit        string            `desc:"Memory hard limit (format: <number>[<unit>], where unit = b, k, m or g)"`
	Metadata           map[string]string `desc:"Metadata to attach to this service"`
	M_Memory           string            `desc:"Memory reservation (format: <number>[<unit>], where unit = b, k, m or g)"`
	Net_Network        string            `desc:"Connect a container to a network (default|home)" default:"default"`
	N_Name             string            `desc:"Assign a name to the container"`
	Permission         []string          `desc:"Permissions to grant to container's service account in current stack"`
	Pid                string            `desc:"PID namespace to use"`
	P_Publish          []string          `desc:"Publish a container's port(s) externally"`
	Privileged         bool              `desc:"Give extended privileges to this container"`
	ReadOnly           bool              `desc:"Mount the container's root filesystem as read only"`
	ReadyCmd           string            `desc:"Command to run to check readiness"`
	ReadyInterval      string            `desc:"Time between running the check (ms|s|m|h)" default:"0s"`
	ReadyRetries       int               `desc:"Consecutive successes needed to report ready"`
	ReadyStartPeriod   string            `desc:"Start period for the container to initialize before starting readychecks (ms|s|m|h)" default:"0s"`
	ReadyTimeout       string            `desc:"Maximum time to allow one check to run (ms|s|m|h)" default:"0s"`
	ReadyURL           string            `desc:"URL to hit to check readiness (example: http://localhost:8080/ping)"`
	Restart            string            `desc:"Restart policy to apply when a container exits" default:"always"`
	Secret             []string          `desc:"Secrets to inject to the service (format: name:target)"`
	SecurityOpt        []string          `desc:"Security Options"`
	StopTimeout        string            `desc:"Timeout (in seconds) to stop a container"`
	Tmpfs              []string          `desc:"Mount a tmpfs directory"`
	T_Tty              bool              `desc:"Allocate a pseudo-TTY"`
	UnhealthyRetries   int               `desc:"Consecutive failures needed to report unhealthy"`
	UnreadyRetries     int               `desc:"Consecutive failures needed to report unready"`
	UpdateOrder        string            `desc:"Update order when doing batched rolling container updates (start-first|stop-first)"`
	UpdateStrategy     string            `desc:"Approach to updating containers (rolling|on-delete)" default:"rolling"`
	U_User             string            `desc:"Username or UID (format: <name|uid>[:<group|gid>])"`
	VolumeDriver       string            `desc:"Optional volume driver for the container"`
	VolumesFrom        []string          `desc:"Mount volumes from the specified container(s)"`
	V_Volume           []string          `desc:"Bind mount a volume"`
	W_Workdir          string            `desc:"Working directory inside the container"`

	Scheduling
}

type Scheduling struct {
	Global         bool     `desc:"Run one container per node (or some nodes depending on scheduling)"`
	Node           string   `desc:"Skip scheduling and run service on specified node"`
	Scheduler      string   `desc:"Use a custom scheduler of the given name"`
	NodeRequire    []string `desc:"Node running containers must match all expressions"`
	NodeRequireAny []string `desc:"Node running containers must match one expression"`
	NodePreferred  []string `desc:"Node running containers if possible should match expression"`
}

func (c *Create) Run(ctx *clicontext.CLIContext) error {
	_, err := c.RunCallback(ctx, func(s *client.Service) *client.Service {
		return s
	})
	return err
}

func (c *Create) RunCallback(ctx *clicontext.CLIContext, cb func(service *client.Service) *client.Service) (*client.Service, error) {
	var err error

	service, err := c.ToService(ctx.CLI.Args())
	if err != nil {
		return nil, err
	}

	service.SpaceID, service.StackID, service.Name, err = stack.ResolveSpaceStackForName(ctx, service.Name)
	if err != nil {
		return nil, err
	}

	wc, err := ctx.WorkspaceClient()
	if err != nil {
		return nil, err
	}

	service = cb(service)

	s, err := wc.Service.Create(service)
	if err != nil {
		return nil, err
	}

	return s, waiter.WaitFor(ctx, &s.Resource)
}

func (c *Create) ToService(args []string) (*client.Service, error) {
	var (
		err error
	)

	if len(args) == 0 {
		return nil, fmt.Errorf("at least one (1) argument is required")
	}

	service := &client.Service{
		CPUs:                c.Cpus,
		CapAdd:              c.CapAdd,
		CapDrop:             c.CapDrop,
		Command:             args[1:],
		DeploymentStrategy:  c.DeploymentStrategy,
		DNS:                 c.Dns,
		DNSOptions:          c.DnsOption,
		DNSSearch:           c.DnsSearch,
		DefaultVolumeDriver: c.VolumeDriver,
		Entrypoint:          c.Entrypoint,
		ExtraHosts:          c.AddHost,
		Global:              c.Global,
		Hostname:            c.Hostname,
		Image:               args[0],
		ImagePullPolicy:     c.ImagePullPolicy,
		Init:                c.Init,
		IpcMode:             c.Ipc,
		Labels:              c.L_Label,
		Name:                c.N_Name,
		NetworkMode:         c.Net_Network,
		OpenStdin:           c.I_Interactive,
		PidMode:             c.Pid,
		Privileged:          c.Privileged,
		ReadonlyRootfs:      c.ReadOnly,
		RestartPolicy:       c.Restart,
		Scheduling: &client.Scheduling{
			Scheduler: c.Scheduler,
			Node: &client.NodeScheduling{
				NodeID:     c.Node,
				RequireAll: c.NodeRequire,
				RequireAny: c.NodeRequireAny,
				Preferred:  c.NodePreferred,
			},
		},
		Tty:            c.T_Tty,
		User:           c.U_User,
		UpdateOrder:    c.UpdateOrder,
		UpdateStrategy: c.UpdateStrategy,
		VolumesFrom:    c.VolumesFrom,
		WorkingDir:     c.W_Workdir,
	}

	service.Volumes, err = ParseMounts(c.V_Volume)
	if err != nil {
		return nil, err
	}

	service.Devices, err = ParseDevices(c.Device)
	if err != nil {
		return nil, err
	}

	service.Configs, err = ParseConfigs(c.Config)
	if err != nil {
		return nil, err
	}

	service.Secrets, err = ParseSecrets(c.Secret)
	if err != nil {
		return nil, err
	}

	service.Metadata = map[string]interface{}{}
	for k, v := range c.Metadata {
		service.Metadata[k] = v
	}

	service.GlobalPermissions, err = ParsePermissions(c.GlobalPermission)
	if err != nil {
		return nil, err
	}

	service.Permissions, err = ParsePermissions(c.Permission)
	if err != nil {
		return nil, err
	}

	service.Environment, err = kvfile.ReadKVEnvStrings(c.EnvFile, c.E_Env)
	if err != nil {
		return nil, err
	}

	service.Labels, err = parseLabels(c.LabelFile, service.Labels)
	if err != nil {
		return nil, err
	}

	if err := populateHealthCheck(c, service); err != nil {
		return nil, err
	}

	if err := populateMemory(c, service); err != nil {
		return nil, err
	}

	service.Tmpfs, err = ParseTmpfs(c.Tmpfs)
	if err != nil {
		return nil, err
	}

	service.PortBindings, err = ParsePorts(c.P_Publish)
	if err != nil {
		return nil, err
	}

	service.ExposedPorts, err = ParseExposedPorts(c.Expose)
	if err != nil {
		return nil, err
	}

	return service, nil
}

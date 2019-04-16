package create

import (
	"fmt"

	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/pretty/stringers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	t = true
)

type Create struct {
	AddHost                []string          `desc:"Add a custom host-to-IP mapping (host:ip)"`
	Annotations            map[string]string `desc:"Annotations to attach to this service"`
	BuildBranch            string            `desc:"Build repository branch" default:"master"`
	BuildSecret            string            `desc:"Set webhook secret"`
	BuildRevision          string            `desc:"Build commit or tag"`
	Command                []string          `desc:"Overwrite the default ENTRYPOINT of the image"`
	Concurrency            int               `desc:"The maximum concurrent request a container can handle(autoscaling)" default:"10"`
	Config                 []string          `desc:"Configs to expose to the service (format: name:target)"`
	Cpus                   string            `desc:"Number of CPUs"`
	DnsOption              []string          `desc:"Set DNS options (format: key:value or key)"`
	DnsSearch              []string          `desc:"Set custom DNS search domains"`
	Dns                    []string          `desc:"Set custom DNS servers"`
	E_Env                  []string          `desc:"Set environment variables"`
	EnvFile                []string          `desc:"Read in a file of environment variables"`
	GlobalPermission       []string          `desc:"Permissions to grant to container's service account for all stacks"`
	Group                  string            `desc:"The GID to run the entrypoint of the container process"`
	HealthCmd              string            `desc:"Command to run to check health"`
	HealthFailureThreshold int               `desc:"Consecutive failures needed to report unhealthy"`
	HealthHeader           map[string]string `desc:"HTTP Headers to send in GET request for healthcheck"`
	HealthInitialDelay     string            `desc:"Start period for the container to initialize before starting healthchecks (ms|s|m|h)" default:"0s"`
	HealthInterval         string            `desc:"Time between running the check (ms|s|m|h)" default:"0s"`
	HealthSuccessThreshold int               `desc:"Consecutive successes needed to report healthy"`
	HealthTimeout          string            `desc:"Maximum time to allow one check to run (ms|s|m|h)" default:"0s"`
	HealthURL              string            `desc:"URL to hit to check health (example: http://localhost:8080/ping)"`
	Hostname               string            `desc:"Container host name"`
	I_Interactive          bool              `desc:"Keep STDIN open even if not attached"`
	ImagePullPolicy        string            `desc:"Behavior determining when to pull the image (never|always|not-present)" default:"not-present"`
	LabelFile              []string          `desc:"Read in a line delimited file of labels"`
	L_Label                map[string]string `desc:"Set meta data on a container"`
	M_Memory               string            `desc:"Memory reservation (format: <number>[<unit>], where unit = b, k, m or g)"`
	N_Name                 string            `desc:"Assign a name to the container"`
	Permission             []string          `desc:"Permissions to grant to container's service account in current stack"`
	P_Ports                []string          `desc:"Publish a container's port(s) externally"`
	ReadOnly               bool              `desc:"Mount the container's root filesystem as read only"`
	Secret                 []string          `desc:"Secrets to inject to the service (format: name:target)"`
	T_Tty                  bool              `desc:"Allocate a pseudo-TTY"`
	U_User                 string            `desc:"UID[:GID] Sets the UID used and optionally GID for entrypoint process (format: <uid>[:<gid>])"`
	W_Workdir              string            `desc:"Working directory inside the container"`
}

func (c *Create) Run(ctx *clicontext.CLIContext) error {
	_, err := c.RunCallback(ctx, func(s *riov1.Service) *riov1.Service {
		return s
	})
	return err
}

func (c *Create) RunCallback(ctx *clicontext.CLIContext, cb func(service *riov1.Service) *riov1.Service) (*riov1.Service, error) {
	var err error

	service, err := c.ToService(ctx.CLI.Args())
	if err != nil {
		return nil, err
	}

	_, service.Namespace, service.Name, err = stack.ResolveSpaceStackForName(ctx, service.Name)
	if err != nil {
		return nil, err
	}

	service = cb(service)

	return service, ctx.Create(service)
}

func (c *Create) ToService(args []string) (*riov1.Service, error) {
	var (
		err error
	)

	if len(args) == 0 {
		return nil, fmt.Errorf("at least one (1) argument is required")
	}

	var spec riov1.ServiceSpec

	spec.Args = args[1:]
	spec.Stdin = c.I_Interactive
	spec.TTY = c.T_Tty
	spec.WorkingDir = c.W_Workdir
	spec.Hostname = c.Hostname
	spec.Nameservers = c.Dns
	spec.Searches = c.DnsSearch

	if c.ReadOnly {
		spec.ReadOnlyRootFilesystem = &t
	}

	spec.ImagePullPolicy, err = stringers.ParseImagePullPolicy(c.ImagePullPolicy)
	if err != nil {
		return nil, err
	}

	spec.HostAliases, err = stringers.ParseHostAliases(c.AddHost...)
	if err != nil {
		return nil, err
	}

	spec.Options = stringers.ParseDNSOptions(c.DnsOption...)

	service := riov1.NewService("", c.N_Name, riov1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      c.L_Label,
			Annotations: c.Annotations,
		},
	})

	spec.CPUs, err = stringers.ParseQuantity(c.Cpus)
	if err != nil {
		return nil, err
	}

	if stringers.IsRepo(args[0]) {
		service.Spec.Build = riov1.ImageBuild{
			Branch:   c.BuildBranch,
			Repo:     args[0],
			Revision: c.BuildRevision,
			Secret:   c.BuildSecret,
		}
	} else {
		service.Spec.Image = args[0]
	}

	spec.RunAsUser, spec.RunAsGroup, err = stringers.ParseUserGroup(c.U_User, c.Group)
	if err != nil {
		return nil, err
	}

	service.Spec.Configs, err = stringers.ParseConfigs(c.Config...)
	if err != nil {
		return nil, err
	}

	service.Spec.Secrets, err = stringers.ParseSecrets(c.Secret...)
	if err != nil {
		return nil, err
	}

	service.Spec.GlobalPermissions, err = stringers.ParsePermissions(c.GlobalPermission...)
	if err != nil {
		return nil, err
	}

	service.Spec.Permissions, err = stringers.ParsePermissions(c.Permission...)
	if err != nil {
		return nil, err
	}

	service.Spec.Env, err = stringers.ParseEnv(c.EnvFile, c.E_Env, true)
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

	service.Spec.Ports, err = stringers.ParsePorts(c.P_Ports...)
	if err != nil {
		return nil, err
	}

	return service, nil
}

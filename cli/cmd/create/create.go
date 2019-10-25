package create

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/stack"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/riofile/stringers"
	"github.com/rancher/wrangler/pkg/kv"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	t = true
)

type Create struct {
	App                    string            `desc:"Specify the app name"`
	AddHost                []string          `desc:"Add a custom host-to-IP mapping (host:ip)"`
	Annotations            map[string]string `desc:"Annotations to attach to this service"`
	BuildBranch            string            `desc:"Build repository branch" default:"master"`
	BuildDockerfile        string            `desc:"Set Dockerfile name, defaults to Dockerfile"`
	BuildContext           string            `desc:"Set build context, defaults to ."`
	BuildWebhookSecret     string            `desc:"Set GitHub webhook secret name"`
	BuildDockerPushSecret  string            `desc:"Set docker push secret name"`
	CloneGitSecret         string            `desc:"Set git clone secret name"`
	BuildImageName         string            `desc:"Specify custom image name to push"`
	BuildRegistry          string            `desc:"Specify to push image to"`
	BuildRevision          string            `desc:"Build git commit or tag"`
	BuildPr                bool              `desc:"Enable pull request builds"`
	BuildTemplate          bool              `desc:"Use this service as a template for generating services if a new commit applys to git repo. If not specified it will do in-place update"`
	BuildTimeout           string            `desc:"Timeout for build, default to 10m (ms|s|m|h)"`
	Command                []string          `desc:"Overwrite the default ENTRYPOINT of the image"`
	Concurrency            int               `desc:"The maximum concurrent request a container can handle (autoscaling)" default:"10"`
	Config                 []string          `desc:"Configs to expose to the service (format: name[/key]:target)"`
	Cpus                   string            `desc:"Number of CPUs"`
	DNSOption              []string          `desc:"Set DNS options (format: key:value or key)"`
	DNSSearch              []string          `desc:"Set custom DNS search domains"`
	DNS                    []string          `desc:"Set custom DNS servers"`
	HostDNS                bool              `desc:"Use the host level DNS and not the cluster level DNS"`
	NoMesh                 bool              `desc:"Disable service mesh"`
	E_Env                  []string          `desc:"Set environment variables"`
	EnvFile                []string          `desc:"Read in a file of environment variables"`
	GlobalPermission       []string          `desc:"Permissions to grant to container's service account for all namespaces"`
	Group                  string            `desc:"The GID to run the entrypoint of the container process"`
	Net                    string            `desc:"Set network mode (host)"`
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
	ImagePullSecrets       []string          `desc:"Specify image pull secrets"`
	LabelFile              []string          `desc:"Read in a line delimited file of labels"`
	L_Label                map[string]string `desc:"Set meta data on a container"`
	M_Memory               string            `desc:"Memory reservation (format: <number>[<unit>], where unit = b, k, m or g)"`
	N_Name                 string            `desc:"Assign a name to the container. Use format ${namespace}/${name} to assign workload to a different namespace"`
	Permission             []string          `desc:"Permissions to grant to container's service account in current namespace"`
	P_Ports                []string          `desc:"Publish a container's port(s) (format: svcport:containerport/protocol)"`
	Privileged             bool              `desc:"Run container with privilege"`
	ReadOnly               bool              `desc:"Mount the container's root filesystem as read only"`
	RolloutInterval        int               `desc:"Rollout interval in seconds"`
	RolloutIncrement       int               `desc:"Rollout increment value"`
	Secret                 []string          `desc:"Secrets to inject to the service (format: name[/key]:target)"`
	Template               bool              `desc:"Use this service as template to rollout services from git"`
	T_Tty                  bool              `desc:"Allocate a pseudo-TTY"`
	Version                string            `desc:"Specify the revision"`
	Scale                  string            `desc:"The number of replicas to run or a range for autoscaling (example 1-10)"`
	U_User                 string            `desc:"UID[:GID] Sets the UID used and optionally GID for entrypoint process (format: <uid>[:<gid>])"`
	Weight                 int               `desc:"Specify the weight for the revision" default:"100"`
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

	service.Namespace, service.Name = stack.NamespaceAndName(ctx, service.Name)

	service = cb(service)
	return service, ctx.Create(service)
}

func (c *Create) setRollout(spec *riov1.ServiceSpec) {
	if c.RolloutIncrement > 0 || c.RolloutInterval > 0 {
		spec.RolloutConfig = &riov1.RolloutConfig{
			Increment: c.RolloutIncrement,
			Interval: metav1.Duration{
				Duration: time.Duration(c.RolloutInterval) * time.Second,
			},
		}
	}
}

func (c *Create) setDNS(spec *riov1.ServiceSpec) (err error) {
	if len(c.DNS) > 0 || len(c.DNSSearch) > 0 || len(c.DNSOption) > 0 || c.HostDNS {
		spec.DNS = &riov1.DNS{
			Nameservers: c.DNS,
			Searches:    c.DNSSearch,
			Options:     nil,
		}
		if c.HostDNS {
			spec.DNS.Policy = v1.DNSDefault
		}
		spec.DNS.Options, err = stringers.ParseDNSOptions(c.DNSOption...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Create) setScale(spec *riov1.ServiceSpec) (err error) {
	if c.Scale == "" {
		return nil
	}

	minStr, maxStr := kv.Split(c.Scale, "-")
	minValue, err := strconv.ParseInt(minStr, 10, 64)
	if err != nil {
		return errors.Wrapf(err, "invalid scale %s", c.Scale)
	}

	min := int32(minValue)
	if maxStr == "" {
		spec.Replicas = &[]int{int(min)}[0]
	} else {
		maxValue, err := strconv.ParseInt(maxStr, 10, 64)
		if err != nil {
			return errors.Wrapf(err, "invalid scale %s", c.Scale)
		}
		max := int32(maxValue)
		spec.Autoscale = &riov1.AutoscaleConfig{
			Concurrency: c.Concurrency,
			MinReplicas: &min,
			MaxReplicas: &[]int32{max}[0],
		}
	}

	return nil
}

func (c *Create) ToService(args []string) (*riov1.Service, error) {
	var (
		err error
	)

	if len(args) == 0 {
		return nil, fmt.Errorf("at least one (1) argument is required")
	}

	var spec riov1.ServiceSpec

	spec.App = c.App
	spec.Args = args[1:]
	spec.Hostname = c.Hostname
	spec.HostNetwork = c.Net == "host"
	spec.Stdin = c.I_Interactive
	spec.TTY = c.T_Tty
	spec.Version = c.Version
	spec.WorkingDir = c.W_Workdir

	if c.NoMesh {
		mesh := !c.NoMesh
		spec.ServiceMesh = &mesh
	}

	if c.Weight > 0 {
		spec.Weight = &c.Weight
	}
	c.setRollout(&spec)
	if err := c.setDNS(&spec); err != nil {
		return nil, err
	}
	if err := c.setScale(&spec); err != nil {
		return nil, err
	}

	spec.ImagePullPolicy, err = stringers.ParseImagePullPolicy(c.ImagePullPolicy)
	if err != nil {
		return nil, err
	}
	spec.ImagePullSecrets = c.ImagePullSecrets

	spec.HostAliases, err = stringers.ParseHostAliases(c.AddHost...)
	if err != nil {
		return nil, err
	}

	if c.Cpus != "" {
		cpus, err := stringers.ParseQuantity(c.Cpus)
		if err != nil {
			return nil, err
		}
		spec.CPUs = &cpus
	}

	if err := c.setBuildOrImage(args[0], &spec); err != nil {
		return nil, err
	}

	spec.ContainerSecurityContext = &riov1.ContainerSecurityContext{}
	if c.ReadOnly {
		spec.ReadOnlyRootFilesystem = &t
	}

	spec.RunAsUser, spec.RunAsGroup, err = stringers.ParseUserGroup(c.U_User, c.Group)
	if err != nil {
		return nil, err
	}

	if c.Privileged {
		spec.Privileged = &c.Privileged
	}

	spec.Configs, err = stringers.ParseConfigs(c.Config...)
	if err != nil {
		return nil, err
	}

	spec.Secrets, err = stringers.ParseSecrets(c.Secret...)
	if err != nil {
		return nil, err
	}

	spec.GlobalPermissions, err = stringers.ParsePermissions(c.GlobalPermission...)
	if err != nil {
		return nil, err
	}

	spec.Permissions, err = stringers.ParsePermissions(c.Permission...)
	if err != nil {
		return nil, err
	}

	spec.Env, err = stringers.ParseAllEnv(c.EnvFile, c.E_Env, true)
	if err != nil {
		return nil, err
	}

	if err := c.setHealthCheck(&spec); err != nil {
		return nil, err
	}

	labels, err := parseLabels(c.LabelFile, c.L_Label)
	if err != nil {
		return nil, err
	}

	if err := c.setMemory(&spec); err != nil {
		return nil, err
	}

	if len(c.P_Ports) == 0 {
		c.P_Ports = []string{"80:8080/http"}
	}

	spec.Ports, err = stringers.ParsePorts(c.P_Ports...)
	if err != nil {
		return nil, err
	}

	return riov1.NewService("", c.N_Name, riov1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: c.Annotations,
		},
		Spec: spec,
	}), nil
}

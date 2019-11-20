package create

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/rancher/rio/cli/cmd/weight"
	"github.com/rancher/rio/cli/pkg/clicontext"
	"github.com/rancher/rio/cli/pkg/types"
	riov1 "github.com/rancher/rio/pkg/apis/rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/riofile/stringers"
	"github.com/rancher/wrangler/pkg/kv"
	"gopkg.in/inf.v0"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	t = true
)

type Create struct {
	AddHost                []string          `desc:"Add a custom host-to-IP mapping (host=ip)"`
	Annotations            map[string]string `desc:"Annotations to attach to this service"`
	BuildBranch            string            `desc:"Build repository branch" default:"master"`
	BuildDockerfile        string            `desc:"Set Dockerfile name, defaults to Dockerfile"`
	BuildContext           string            `desc:"Set build context, defaults to ."`
	BuildWebhookSecret     string            `desc:"Set GitHub webhook secret name"`
	BuildDockerPushSecret  string            `desc:"Set docker push secret name"`
	BuildCloneSecret       string            `desc:"Set git clone secret name"`
	BuildImageName         string            `desc:"Specify custom image name to push"`
	BuildRegistry          string            `desc:"Specify to push image to"`
	BuildRevision          string            `desc:"Build git commit or tag"`
	BuildPr                bool              `desc:"Enable pull request builds"`
	BuildTimeout           string            `desc:"Timeout for build, default to 10m (ms|s|m|h)"`
	Command                []string          `desc:"Overwrite the default ENTRYPOINT of the image"`
	Config                 []string          `desc:"Configs to expose to the service (format: name[/key]:target)"`
	Concurrency            int               `desc:"The maximum concurrent request a container can handle (autoscaling)" default:"10"`
	Cpus                   string            `desc:"Number of CPUs"`
	DNS                    []string          `desc:"Set custom DNS servers"`
	DNSOption              []string          `desc:"Set DNS options (format: key:value or key)"`
	DNSSearch              []string          `desc:"Set custom DNS search domains"`
	E_Env                  []string          `desc:"Set environment variables"`
	EnvFile                []string          `desc:"Read in a file of environment variables"`
	GlobalPermission       []string          `desc:"Permissions to grant to container's service account for all namespaces"`
	Group                  string            `desc:"The GID to run the entrypoint of the container process"`
	HealthCmd              string            `desc:"Command to run to check health"`
	HealthFailureThreshold int               `desc:"Consecutive failures needed to report unhealthy"`
	HealthHeader           map[string]string `desc:"HTTP Headers to send in GET request for healthcheck"`
	HealthInitialDelay     string            `desc:"Start period for the container to initialize before starting healthchecks (ms|s|m|h)" default:"0s"`
	HealthInterval         string            `desc:"Time between running the check (ms|s|m|h)" default:"0s"`
	HealthSuccessThreshold int               `desc:"Consecutive successes needed to report healthy"`
	HealthTimeout          string            `desc:"Maximum time to allow one check to run (ms|s|m|h)" default:"0s"`
	HealthURL              string            `desc:"URL to hit to check health (example: http://:8080/ping)"`
	HostDNS                bool              `desc:"Use the host level DNS and not the cluster level DNS"`
	Hostname               string            `desc:"Container host name"`
	ImagePullPolicy        string            `desc:"Behavior determining when to pull the image (never|always|not-present)" default:"not-present"`
	ImagePullSecrets       []string          `desc:"Specify image pull secrets"`
	I_Interactive          bool              `desc:"Keep STDIN open even if not attached"`
	LabelFile              []string          `desc:"Read in a line delimited file of labels"`
	L_Label                map[string]string `desc:"Set meta data on a container"`
	M_Memory               string            `desc:"Memory reservation (format: <number>[<unit>], where unit = b, k, m or g)"`
	N_Name                 string            `desc:"Assign a name to the container. Use format [namespace:]name[@version]"`
	Net                    string            `desc:"Set network mode (host)"`
	NoMesh                 bool              `desc:"Disable service mesh"`
	Permission             []string          `desc:"Permissions to grant to container's service account in current namespace"`
	P_Ports                []string          `desc:"Publish a container's port(s) (format: svcport:containerport/protocol)"`
	Privileged             bool              `desc:"Run container with privilege"`
	ReadOnly               bool              `desc:"Mount the container's root filesystem as read only"`
	RolloutDuration        string            `desc:"How long the rollout should take" default:"0s"`
	RequestTimeoutSeconds  int               `desc:"Set request timeout in seconds"`
	Scale                  string            `desc:"The number of replicas to run or a range for autoscaling (example 1-10)"`
	Secret                 []string          `desc:"Secrets to inject to the service (format: name[/key]:target)"`
	StageOnly              bool              `desc:"Only stage service when generating new services. Can only be used when template is true"`
	Template               bool              `desc:"If true new version is created per git commit. If false update in-place"`
	T_Tty                  bool              `desc:"Allocate a pseudo-TTY"`
	U_User                 string            `desc:"UID[:GID] Sets the UID used and optionally GID for entrypoint process (format: <uid>[:<gid>])"`
	V_Volume               []string          `desc:"Specify volumes for for services"`
	Weight                 int               `desc:"Specify the weight for the services"`
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

	service, err := c.ToService(ctx, ctx.CLI.Args())
	if err != nil {
		return nil, err
	}

	service = cb(service)
	err = ctx.Create(service)
	if err != nil {
		return nil, err
	}
	if c.Weight == 100 {
		err = weight.PromoteService(ctx, types.Resource{
			Name:      service.Name,
			App:       service.Spec.App,
			Version:   service.Spec.Version,
			Namespace: service.Namespace,
		}, service.Spec.RolloutConfig, *service.Spec.Weight)
		if err != nil {
			return service, err
		}
	}
	return service, nil
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

func (c *Create) ToService(ctx *clicontext.CLIContext, args []string) (*riov1.Service, error) {
	var (
		err error
	)

	if len(args) == 0 {
		return nil, fmt.Errorf("at least one (1) argument is required")
	}

	name := c.N_Name
	if name == "" {
		name = clicontext.RandomName()
	}

	r := ctx.ParseID(name)

	if _, err := ctx.GetService(r); err == nil {
		return nil, fmt.Errorf("%s already exists", r)
	}

	var spec riov1.ServiceSpec

	spec.App = r.App
	spec.Args = args[1:]
	spec.Hostname = c.Hostname
	spec.HostNetwork = c.Net == "host"
	spec.Stdin = c.I_Interactive
	spec.TTY = c.T_Tty
	spec.Version = r.Version
	spec.WorkingDir = c.W_Workdir
	spec.Template = c.Template
	spec.StageOnly = c.StageOnly

	if c.NoMesh {
		mesh := !c.NoMesh
		spec.ServiceMesh = &mesh
	}

	if c.Weight > 100 {
		return nil, fmt.Errorf("weight cannot exceed 100")
	}

	if c.Weight > 0 {
		duration, err := time.ParseDuration(c.RolloutDuration)
		if err != nil {
			return nil, err
		}
		tempResource := types.Resource{
			Name:      name,
			App:       r.App,
			Version:   r.Version,
			Namespace: r.Namespace,
		}
		newWeight, rc, err := weight.GenerateWeightAndRolloutConfig(ctx, tempResource, c.Weight, duration, false)
		if err != nil {
			return nil, err
		}
		spec.Weight = &newWeight
		spec.RolloutConfig = rc
	}

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
	if spec.ImagePullPolicy == v1.PullIfNotPresent {
		spec.ImagePullPolicy = ""
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

		v := cpus.ToDec().AsDec()
		cpuMillis := v.Mul(v, inf.NewDec(1000, 1)).UnscaledBig().Int64()
		spec.CPUMillis = &cpuMillis
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

	spec.Volumes, err = stringers.ParseVolumes(c.V_Volume...)
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

	if c.RequestTimeoutSeconds != 0 {
		spec.RequestTimeoutSeconds = &c.RequestTimeoutSeconds
	}

	if err := c.setMemory(&spec); err != nil {
		return nil, err
	}

	spec.Ports, err = stringers.ParsePorts(c.P_Ports...)
	if err != nil {
		return nil, err
	}

	svc := riov1.NewService(r.Namespace, "", riov1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      labels,
			Annotations: c.Annotations,
		},
		Spec: spec,
	})
	svc.GenerateName = spec.App + "-" + spec.Version

	return svc, nil
}

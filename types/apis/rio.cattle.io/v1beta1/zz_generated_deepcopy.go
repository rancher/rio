package v1beta1

import (
	v3 "github.com/rancher/types/apis/management.cattle.io/v3"
	v1beta2 "k8s.io/api/apps/v1beta2"
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Abort) DeepCopyInto(out *Abort) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Abort.
func (in *Abort) DeepCopy() *Abort {
	if in == nil {
		return nil
	}
	out := new(Abort)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BindOptions) DeepCopyInto(out *BindOptions) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BindOptions.
func (in *BindOptions) DeepCopy() *BindOptions {
	if in == nil {
		return nil
	}
	out := new(BindOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Condition) DeepCopyInto(out *Condition) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Condition.
func (in *Condition) DeepCopy() *Condition {
	if in == nil {
		return nil
	}
	out := new(Condition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Config) DeepCopyInto(out *Config) {
	*out = *in
	out.Namespaced = in.Namespaced
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Config.
func (in *Config) DeepCopy() *Config {
	if in == nil {
		return nil
	}
	out := new(Config)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Config) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConfigList) DeepCopyInto(out *ConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Config, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConfigList.
func (in *ConfigList) DeepCopy() *ConfigList {
	if in == nil {
		return nil
	}
	out := new(ConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConfigMapping) DeepCopyInto(out *ConfigMapping) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConfigMapping.
func (in *ConfigMapping) DeepCopy() *ConfigMapping {
	if in == nil {
		return nil
	}
	out := new(ConfigMapping)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConfigSpec) DeepCopyInto(out *ConfigSpec) {
	*out = *in
	out.StackScoped = in.StackScoped
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConfigSpec.
func (in *ConfigSpec) DeepCopy() *ConfigSpec {
	if in == nil {
		return nil
	}
	out := new(ConfigSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ContainerConfig) DeepCopyInto(out *ContainerConfig) {
	*out = *in
	out.ContainerPrivilegedConfig = in.ContainerPrivilegedConfig
	if in.CapAdd != nil {
		in, out := &in.CapAdd, &out.CapAdd
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.CapDrop != nil {
		in, out := &in.CapDrop, &out.CapDrop
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Command != nil {
		in, out := &in.Command, &out.Command
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Entrypoint != nil {
		in, out := &in.Entrypoint, &out.Entrypoint
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Environment != nil {
		in, out := &in.Environment, &out.Environment
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ExposedPorts != nil {
		in, out := &in.ExposedPorts, &out.ExposedPorts
		*out = make([]ExposedPort, len(*in))
		copy(*out, *in)
	}
	if in.Healthcheck != nil {
		in, out := &in.Healthcheck, &out.Healthcheck
		if *in == nil {
			*out = nil
		} else {
			*out = new(HealthConfig)
			(*in).DeepCopyInto(*out)
		}
	}
	if in.Readycheck != nil {
		in, out := &in.Readycheck, &out.Readycheck
		if *in == nil {
			*out = nil
		} else {
			*out = new(HealthConfig)
			(*in).DeepCopyInto(*out)
		}
	}
	if in.PortBindings != nil {
		in, out := &in.PortBindings, &out.PortBindings
		*out = make([]PortBinding, len(*in))
		copy(*out, *in)
	}
	if in.Tmpfs != nil {
		in, out := &in.Tmpfs, &out.Tmpfs
		*out = make([]Tmpfs, len(*in))
		copy(*out, *in)
	}
	if in.Volumes != nil {
		in, out := &in.Volumes, &out.Volumes
		*out = make([]Mount, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.VolumesFrom != nil {
		in, out := &in.VolumesFrom, &out.VolumesFrom
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Devices != nil {
		in, out := &in.Devices, &out.Devices
		*out = make([]DeviceMapping, len(*in))
		copy(*out, *in)
	}
	if in.Configs != nil {
		in, out := &in.Configs, &out.Configs
		*out = make([]ConfigMapping, len(*in))
		copy(*out, *in)
	}
	if in.Secrets != nil {
		in, out := &in.Secrets, &out.Secrets
		*out = make([]SecretMapping, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ContainerConfig.
func (in *ContainerConfig) DeepCopy() *ContainerConfig {
	if in == nil {
		return nil
	}
	out := new(ContainerConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ContainerPrivilegedConfig) DeepCopyInto(out *ContainerPrivilegedConfig) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ContainerPrivilegedConfig.
func (in *ContainerPrivilegedConfig) DeepCopy() *ContainerPrivilegedConfig {
	if in == nil {
		return nil
	}
	out := new(ContainerPrivilegedConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CustomResourceDefinition) DeepCopyInto(out *CustomResourceDefinition) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CustomResourceDefinition.
func (in *CustomResourceDefinition) DeepCopy() *CustomResourceDefinition {
	if in == nil {
		return nil
	}
	out := new(CustomResourceDefinition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Destination) DeepCopyInto(out *Destination) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Destination.
func (in *Destination) DeepCopy() *Destination {
	if in == nil {
		return nil
	}
	out := new(Destination)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DeviceMapping) DeepCopyInto(out *DeviceMapping) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DeviceMapping.
func (in *DeviceMapping) DeepCopy() *DeviceMapping {
	if in == nil {
		return nil
	}
	out := new(DeviceMapping)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExposedPort) DeepCopyInto(out *ExposedPort) {
	*out = *in
	out.PortBinding = in.PortBinding
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExposedPort.
func (in *ExposedPort) DeepCopy() *ExposedPort {
	if in == nil {
		return nil
	}
	out := new(ExposedPort)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Fault) DeepCopyInto(out *Fault) {
	*out = *in
	out.Abort = in.Abort
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Fault.
func (in *Fault) DeepCopy() *Fault {
	if in == nil {
		return nil
	}
	out := new(Fault)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HealthConfig) DeepCopyInto(out *HealthConfig) {
	*out = *in
	if in.Test != nil {
		in, out := &in.Test, &out.Test
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HealthConfig.
func (in *HealthConfig) DeepCopy() *HealthConfig {
	if in == nil {
		return nil
	}
	out := new(HealthConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InternalStack) DeepCopyInto(out *InternalStack) {
	*out = *in
	if in.Services != nil {
		in, out := &in.Services, &out.Services
		*out = make(map[string]Service, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.Configs != nil {
		in, out := &in.Configs, &out.Configs
		*out = make(map[string]Config, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.Volumes != nil {
		in, out := &in.Volumes, &out.Volumes
		*out = make(map[string]Volume, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.Routes != nil {
		in, out := &in.Routes, &out.Routes
		*out = make(map[string]RouteSet, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	in.Kubernetes.DeepCopyInto(&out.Kubernetes)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InternalStack.
func (in *InternalStack) DeepCopy() *InternalStack {
	if in == nil {
		return nil
	}
	out := new(InternalStack)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Kubernetes) DeepCopyInto(out *Kubernetes) {
	*out = *in
	if in.CustomResourceDefinitions != nil {
		in, out := &in.CustomResourceDefinitions, &out.CustomResourceDefinitions
		*out = make([]CustomResourceDefinition, len(*in))
		copy(*out, *in)
	}
	if in.NamespacedCustomResourceDefinitions != nil {
		in, out := &in.NamespacedCustomResourceDefinitions, &out.NamespacedCustomResourceDefinitions
		*out = make([]CustomResourceDefinition, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Kubernetes.
func (in *Kubernetes) DeepCopy() *Kubernetes {
	if in == nil {
		return nil
	}
	out := new(Kubernetes)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Match) DeepCopyInto(out *Match) {
	*out = *in
	out.Path = in.Path
	out.Scheme = in.Scheme
	out.Method = in.Method
	if in.Headers != nil {
		in, out := &in.Headers, &out.Headers
		*out = make(map[string]StringMatch, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Cookies != nil {
		in, out := &in.Cookies, &out.Cookies
		*out = make(map[string]StringMatch, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	out.From = in.From
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Match.
func (in *Match) DeepCopy() *Match {
	if in == nil {
		return nil
	}
	out := new(Match)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Mount) DeepCopyInto(out *Mount) {
	*out = *in
	if in.BindOptions != nil {
		in, out := &in.BindOptions, &out.BindOptions
		if *in == nil {
			*out = nil
		} else {
			*out = new(BindOptions)
			**out = **in
		}
	}
	if in.VolumeOptions != nil {
		in, out := &in.VolumeOptions, &out.VolumeOptions
		if *in == nil {
			*out = nil
		} else {
			*out = new(VolumeOptions)
			**out = **in
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Mount.
func (in *Mount) DeepCopy() *Mount {
	if in == nil {
		return nil
	}
	out := new(Mount)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NodeScheduling) DeepCopyInto(out *NodeScheduling) {
	*out = *in
	if in.RequireAll != nil {
		in, out := &in.RequireAll, &out.RequireAll
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.RequireAny != nil {
		in, out := &in.RequireAny, &out.RequireAny
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Preferred != nil {
		in, out := &in.Preferred, &out.Preferred
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NodeScheduling.
func (in *NodeScheduling) DeepCopy() *NodeScheduling {
	if in == nil {
		return nil
	}
	out := new(NodeScheduling)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Permission) DeepCopyInto(out *Permission) {
	*out = *in
	if in.Verbs != nil {
		in, out := &in.Verbs, &out.Verbs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Permission.
func (in *Permission) DeepCopy() *Permission {
	if in == nil {
		return nil
	}
	out := new(Permission)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodConfig) DeepCopyInto(out *PodConfig) {
	*out = *in
	in.Scheduling.DeepCopyInto(&out.Scheduling)
	if in.StopGracePeriodSeconds != nil {
		in, out := &in.StopGracePeriodSeconds, &out.StopGracePeriodSeconds
		if *in == nil {
			*out = nil
		} else {
			*out = new(int)
			**out = **in
		}
	}
	if in.DNS != nil {
		in, out := &in.DNS, &out.DNS
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.DNSOptions != nil {
		in, out := &in.DNSOptions, &out.DNSOptions
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.DNSSearch != nil {
		in, out := &in.DNSSearch, &out.DNSSearch
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ExtraHosts != nil {
		in, out := &in.ExtraHosts, &out.ExtraHosts
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.GlobalPermissions != nil {
		in, out := &in.GlobalPermissions, &out.GlobalPermissions
		*out = make([]Permission, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Permissions != nil {
		in, out := &in.Permissions, &out.Permissions
		*out = make([]Permission, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodConfig.
func (in *PodConfig) DeepCopy() *PodConfig {
	if in == nil {
		return nil
	}
	out := new(PodConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PortBinding) DeepCopyInto(out *PortBinding) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PortBinding.
func (in *PortBinding) DeepCopy() *PortBinding {
	if in == nil {
		return nil
	}
	out := new(PortBinding)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PrivilegedConfig) DeepCopyInto(out *PrivilegedConfig) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PrivilegedConfig.
func (in *PrivilegedConfig) DeepCopy() *PrivilegedConfig {
	if in == nil {
		return nil
	}
	out := new(PrivilegedConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Redirect) DeepCopyInto(out *Redirect) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Redirect.
func (in *Redirect) DeepCopy() *Redirect {
	if in == nil {
		return nil
	}
	out := new(Redirect)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Retry) DeepCopyInto(out *Retry) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Retry.
func (in *Retry) DeepCopy() *Retry {
	if in == nil {
		return nil
	}
	out := new(Retry)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Rewrite) DeepCopyInto(out *Rewrite) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Rewrite.
func (in *Rewrite) DeepCopy() *Rewrite {
	if in == nil {
		return nil
	}
	out := new(Rewrite)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RouteSet) DeepCopyInto(out *RouteSet) {
	*out = *in
	out.Namespaced = in.Namespaced
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouteSet.
func (in *RouteSet) DeepCopy() *RouteSet {
	if in == nil {
		return nil
	}
	out := new(RouteSet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RouteSet) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RouteSetList) DeepCopyInto(out *RouteSetList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RouteSet, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouteSetList.
func (in *RouteSetList) DeepCopy() *RouteSetList {
	if in == nil {
		return nil
	}
	out := new(RouteSetList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RouteSetList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RouteSetSpec) DeepCopyInto(out *RouteSetSpec) {
	*out = *in
	if in.Routes != nil {
		in, out := &in.Routes, &out.Routes
		*out = make([]RouteSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	out.StackScoped = in.StackScoped
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouteSetSpec.
func (in *RouteSetSpec) DeepCopy() *RouteSetSpec {
	if in == nil {
		return nil
	}
	out := new(RouteSetSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RouteSpec) DeepCopyInto(out *RouteSpec) {
	*out = *in
	if in.Matches != nil {
		in, out := &in.Matches, &out.Matches
		*out = make([]Match, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.To != nil {
		in, out := &in.To, &out.To
		*out = make([]WeightedDestination, len(*in))
		copy(*out, *in)
	}
	out.Redirect = in.Redirect
	out.Rewrite = in.Rewrite
	if in.AddHeaders != nil {
		in, out := &in.AddHeaders, &out.AddHeaders
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	out.RouteTraffic = in.RouteTraffic
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouteSpec.
func (in *RouteSpec) DeepCopy() *RouteSpec {
	if in == nil {
		return nil
	}
	out := new(RouteSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RouteTraffic) DeepCopyInto(out *RouteTraffic) {
	*out = *in
	out.Fault = in.Fault
	out.Mirror = in.Mirror
	out.Retry = in.Retry
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouteTraffic.
func (in *RouteTraffic) DeepCopy() *RouteTraffic {
	if in == nil {
		return nil
	}
	out := new(RouteTraffic)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ScaleStatus) DeepCopyInto(out *ScaleStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ScaleStatus.
func (in *ScaleStatus) DeepCopy() *ScaleStatus {
	if in == nil {
		return nil
	}
	out := new(ScaleStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Scheduling) DeepCopyInto(out *Scheduling) {
	*out = *in
	in.Node.DeepCopyInto(&out.Node)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Scheduling.
func (in *Scheduling) DeepCopy() *Scheduling {
	if in == nil {
		return nil
	}
	out := new(Scheduling)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretMapping) DeepCopyInto(out *SecretMapping) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretMapping.
func (in *SecretMapping) DeepCopy() *SecretMapping {
	if in == nil {
		return nil
	}
	out := new(SecretMapping)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Service) DeepCopyInto(out *Service) {
	*out = *in
	out.Namespaced = in.Namespaced
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Service.
func (in *Service) DeepCopy() *Service {
	if in == nil {
		return nil
	}
	out := new(Service)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Service) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceList) DeepCopyInto(out *ServiceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Service, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceList.
func (in *ServiceList) DeepCopy() *ServiceList {
	if in == nil {
		return nil
	}
	out := new(ServiceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceRevision) DeepCopyInto(out *ServiceRevision) {
	*out = *in
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceRevision.
func (in *ServiceRevision) DeepCopy() *ServiceRevision {
	if in == nil {
		return nil
	}
	out := new(ServiceRevision)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceSource) DeepCopyInto(out *ServiceSource) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceSource.
func (in *ServiceSource) DeepCopy() *ServiceSource {
	if in == nil {
		return nil
	}
	out := new(ServiceSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceSpec) DeepCopyInto(out *ServiceSpec) {
	*out = *in
	in.ServiceUnversionedSpec.DeepCopyInto(&out.ServiceUnversionedSpec)
	out.StackScoped = in.StackScoped
	if in.Revisions != nil {
		in, out := &in.Revisions, &out.Revisions
		*out = make(map[string]ServiceRevision, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceSpec.
func (in *ServiceSpec) DeepCopy() *ServiceSpec {
	if in == nil {
		return nil
	}
	out := new(ServiceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceStatus) DeepCopyInto(out *ServiceStatus) {
	*out = *in
	if in.DeploymentStatus != nil {
		in, out := &in.DeploymentStatus, &out.DeploymentStatus
		if *in == nil {
			*out = nil
		} else {
			*out = new(v1beta2.DeploymentStatus)
			(*in).DeepCopyInto(*out)
		}
	}
	if in.ScaleStatus != nil {
		in, out := &in.ScaleStatus, &out.ScaleStatus
		if *in == nil {
			*out = nil
		} else {
			*out = new(ScaleStatus)
			**out = **in
		}
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]Condition, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceStatus.
func (in *ServiceStatus) DeepCopy() *ServiceStatus {
	if in == nil {
		return nil
	}
	out := new(ServiceStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceUnversionedSpec) DeepCopyInto(out *ServiceUnversionedSpec) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Metadata != nil {
		in, out := &in.Metadata, &out.Metadata
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.PodConfig.DeepCopyInto(&out.PodConfig)
	out.PrivilegedConfig = in.PrivilegedConfig
	if in.Sidekicks != nil {
		in, out := &in.Sidekicks, &out.Sidekicks
		*out = make(map[string]SidekickConfig, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	in.ContainerConfig.DeepCopyInto(&out.ContainerConfig)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceUnversionedSpec.
func (in *ServiceUnversionedSpec) DeepCopy() *ServiceUnversionedSpec {
	if in == nil {
		return nil
	}
	out := new(ServiceUnversionedSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SidekickConfig) DeepCopyInto(out *SidekickConfig) {
	*out = *in
	in.ContainerConfig.DeepCopyInto(&out.ContainerConfig)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SidekickConfig.
func (in *SidekickConfig) DeepCopy() *SidekickConfig {
	if in == nil {
		return nil
	}
	out := new(SidekickConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Stack) DeepCopyInto(out *Stack) {
	*out = *in
	out.Namespaced = in.Namespaced
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Stack.
func (in *Stack) DeepCopy() *Stack {
	if in == nil {
		return nil
	}
	out := new(Stack)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Stack) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackList) DeepCopyInto(out *StackList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Stack, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackList.
func (in *StackList) DeepCopy() *StackList {
	if in == nil {
		return nil
	}
	out := new(StackList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *StackList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackScoped) DeepCopyInto(out *StackScoped) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackScoped.
func (in *StackScoped) DeepCopy() *StackScoped {
	if in == nil {
		return nil
	}
	out := new(StackScoped)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackSpec) DeepCopyInto(out *StackSpec) {
	*out = *in
	if in.AdditionalFiles != nil {
		in, out := &in.AdditionalFiles, &out.AdditionalFiles
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Answers != nil {
		in, out := &in.Answers, &out.Answers
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Questions != nil {
		in, out := &in.Questions, &out.Questions
		*out = make([]v3.Question, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.RepoTag != nil {
		in, out := &in.RepoTag, &out.RepoTag
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackSpec.
func (in *StackSpec) DeepCopy() *StackSpec {
	if in == nil {
		return nil
	}
	out := new(StackSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StackStatus) DeepCopyInto(out *StackStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]Condition, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StackStatus.
func (in *StackStatus) DeepCopy() *StackStatus {
	if in == nil {
		return nil
	}
	out := new(StackStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StringMatch) DeepCopyInto(out *StringMatch) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StringMatch.
func (in *StringMatch) DeepCopy() *StringMatch {
	if in == nil {
		return nil
	}
	out := new(StringMatch)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Tmpfs) DeepCopyInto(out *Tmpfs) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Tmpfs.
func (in *Tmpfs) DeepCopy() *Tmpfs {
	if in == nil {
		return nil
	}
	out := new(Tmpfs)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Volume) DeepCopyInto(out *Volume) {
	*out = *in
	out.Namespaced = in.Namespaced
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Volume.
func (in *Volume) DeepCopy() *Volume {
	if in == nil {
		return nil
	}
	out := new(Volume)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Volume) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeList) DeepCopyInto(out *VolumeList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Volume, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeList.
func (in *VolumeList) DeepCopy() *VolumeList {
	if in == nil {
		return nil
	}
	out := new(VolumeList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *VolumeList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeOptions) DeepCopyInto(out *VolumeOptions) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeOptions.
func (in *VolumeOptions) DeepCopy() *VolumeOptions {
	if in == nil {
		return nil
	}
	out := new(VolumeOptions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeSpec) DeepCopyInto(out *VolumeSpec) {
	*out = *in
	out.StackScoped = in.StackScoped
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeSpec.
func (in *VolumeSpec) DeepCopy() *VolumeSpec {
	if in == nil {
		return nil
	}
	out := new(VolumeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VolumeStatus) DeepCopyInto(out *VolumeStatus) {
	*out = *in
	if in.PVCStatus != nil {
		in, out := &in.PVCStatus, &out.PVCStatus
		if *in == nil {
			*out = nil
		} else {
			*out = new(v1.PersistentVolumeClaimStatus)
			(*in).DeepCopyInto(*out)
		}
	}
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]Condition, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VolumeStatus.
func (in *VolumeStatus) DeepCopy() *VolumeStatus {
	if in == nil {
		return nil
	}
	out := new(VolumeStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WeightedDestination) DeepCopyInto(out *WeightedDestination) {
	*out = *in
	out.Destination = in.Destination
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WeightedDestination.
func (in *WeightedDestination) DeepCopy() *WeightedDestination {
	if in == nil {
		return nil
	}
	out := new(WeightedDestination)
	in.DeepCopyInto(out)
	return out
}

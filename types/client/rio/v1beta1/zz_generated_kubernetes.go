package client

const (
	KubernetesType                                     = "kubernetes"
	KubernetesFieldCustomResourceDefinitions           = "customResourceDefinitions"
	KubernetesFieldManifest                            = "manifest"
	KubernetesFieldNamespacedCustomResourceDefinitions = "namespacedCustomResourceDefinitions"
	KubernetesFieldNamespacedManifest                  = "namespacedManifest"
)

type Kubernetes struct {
	CustomResourceDefinitions           []CustomResourceDefinition `json:"customResourceDefinitions,omitempty" yaml:"customResourceDefinitions,omitempty"`
	Manifest                            string                     `json:"manifest,omitempty" yaml:"manifest,omitempty"`
	NamespacedCustomResourceDefinitions []CustomResourceDefinition `json:"namespacedCustomResourceDefinitions,omitempty" yaml:"namespacedCustomResourceDefinitions,omitempty"`
	NamespacedManifest                  string                     `json:"namespacedManifest,omitempty" yaml:"namespacedManifest,omitempty"`
}

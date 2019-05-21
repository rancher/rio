package config

import (
	"bytes"
	"os"
	"text/template"

	"github.com/rancher/rio/pkg/constants"
	"istio.io/api/mesh/v1alpha1"
	"istio.io/istio/pilot/pkg/kube/inject"
	"istio.io/istio/pilot/pkg/model"
	v1 "k8s.io/api/core/v1"
)

func InjectParams(meshConfig *v1alpha1.MeshConfig) *inject.Params {
	debug := false
	os.Setenv("ISTIO_PROXY_IMAGE", "proxyv2")
	hub := "docker.io/istio"
	tag := "1.1.3"

	return &inject.Params{
		InitImage:           inject.InitImageName(hub, tag, false),
		ProxyImage:          inject.ProxyImageName(hub, tag, false),
		Verbosity:           2,
		SidecarProxyUID:     uint64(1337),
		Version:             "",
		EnableCoreDump:      debug,
		Mesh:                meshConfig,
		ImagePullPolicy:     string(v1.PullAlways),
		IncludeIPRanges:     constants.ServiceCidr,
		ExcludeIPRanges:     "10.43.0.0/31",
		IncludeInboundPorts: "*",
		ExcludeInboundPorts: "",
		DebugMode:           debug,
		StatusPort:          15020,
	}
}

func DoConfigAndTemplate(config, templates string) (*v1alpha1.MeshConfig, string, error) {
	meshConfig, err := model.ApplyMeshConfigDefaults(config)
	if err != nil {
		return nil, "", err
	}

	params := InjectParams(meshConfig)
	content, err := generateTemplateFromParams(params, templates)

	return meshConfig, content, err
}

func generateTemplateFromParams(params *inject.Params, templates string) (string, error) {
	// Validate the parameters before we go any farther.
	if err := params.Validate(); err != nil {
		return "", err
	}

	var tmp bytes.Buffer
	err := template.Must(template.New("inject").Parse(templates)).Execute(&tmp, params)
	return tmp.String(), err
}

func Inject(input []byte, template string, meshConfig *v1alpha1.MeshConfig) ([]byte, error) {
	in := bytes.NewBuffer(input)
	out := bytes.NewBuffer(make([]byte, 0, len(input)))
	err := inject.IntoResourceFile(template, meshConfig, in, out)
	return out.Bytes(), err
}

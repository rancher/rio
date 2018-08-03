package config

import (
	"bytes"

	"os"

	"istio.io/api/mesh/v1alpha1"
	"istio.io/istio/pilot/pkg/kube/inject"
	"istio.io/istio/pilot/pkg/model"
	"k8s.io/api/core/v1"
)

func InjectParams(meshConfig *v1alpha1.MeshConfig) *inject.Params {
	debug := true
	os.Setenv("ISTIO_PROXY_IMAGE", "proxyv2")
	hub := "docker.io/istio"
	tag := "0.8.0"

	return &inject.Params{
		InitImage:           inject.InitImageName(hub, tag, debug),
		ProxyImage:          inject.ProxyImageName(hub, tag, debug),
		Verbosity:           2,
		SidecarProxyUID:     uint64(1337),
		Version:             "",
		EnableCoreDump:      debug,
		Mesh:                meshConfig,
		ImagePullPolicy:     string(v1.PullIfNotPresent),
		IncludeIPRanges:     "",
		ExcludeIPRanges:     "",
		IncludeInboundPorts: "*",
		ExcludeInboundPorts: "",
		DebugMode:           debug,
	}
}

func ConfigAndTemplate(config string) (*v1alpha1.MeshConfig, string, error) {
	meshConfig, err := model.ApplyMeshConfigDefaults(config)
	if err != nil {
		return nil, "", err
	}

	params := InjectParams(meshConfig)
	template, err := inject.GenerateTemplateFromParams(params)
	if err != nil {
		return nil, "", err
	}
	return meshConfig, template, nil
}

func Inject(input []byte, template string, meshConfig *v1alpha1.MeshConfig) ([]byte, error) {
	in := bytes.NewBuffer(input)
	out := bytes.NewBuffer(make([]byte, 0, len(input)))
	err := inject.IntoResourceFile(template, meshConfig, in, out)
	return out.Bytes(), err
}

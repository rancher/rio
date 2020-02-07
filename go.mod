module github.com/rancher/rio

go 1.13

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.2.0+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.1
	github.com/containerd/containerd v1.3.0-0.20190507210959-7c1e88399ec0 => github.com/containerd/containerd v1.2.1-0.20190507210959-7c1e88399ec0
	// needed for containerd
	github.com/docker/distribution => github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible
	github.com/docker/docker v1.14.0-0.20190319215453-e7b5f7dbe98c => github.com/docker/docker v1.4.2-0.20190319215453-e7b5f7dbe98c
	// vt100 needed for buildkit
	github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305
	github.com/jetstack/cert-manager v0.11.0 => github.com/rancher/cert-manager v0.5.1-0.20191021233300-3a070253aeda
	github.com/linkerd/linkerd2 => github.com/StrongMonkey/linkerd2 v0.0.0-20191021165729-976fad67457a
	github.com/pseudomuto/protoc-gen-doc => github.com/pseudomuto/protoc-gen-doc v1.0.0
	github.com/wercker/stern v1.11.0 => github.com/rancher/stern v0.0.0-20191213223518-59c2bf84f705
	golang.org/x/crypto v0.0.0-20190129210102-0709b304e793 => golang.org/x/crypto v0.0.0-20180904163835-0709b304e793
	k8s.io/api => github.com/rancher/kubernetes/staging/src/k8s.io/api v1.16.2-k3s.1
	k8s.io/apiextensions-apiserver => github.com/rancher/kubernetes/staging/src/k8s.io/apiextensions-apiserver v1.16.2-k3s.1
	k8s.io/apimachinery => github.com/rancher/kubernetes/staging/src/k8s.io/apimachinery v1.16.2-k3s.1
	k8s.io/apiserver => github.com/rancher/kubernetes/staging/src/k8s.io/apiserver v1.16.2-k3s.1
	k8s.io/cli-runtime => github.com/rancher/kubernetes/staging/src/k8s.io/cli-runtime v1.16.2-k3s.1
	k8s.io/client-go => github.com/rancher/kubernetes/staging/src/k8s.io/client-go v1.16.2-k3s.1
	k8s.io/cloud-provider => github.com/rancher/kubernetes/staging/src/k8s.io/cloud-provider v1.16.2-k3s.1
	k8s.io/cluster-bootstrap => github.com/rancher/kubernetes/staging/src/k8s.io/cluster-bootstrap v1.16.2-k3s.1
	k8s.io/code-generator => github.com/rancher/kubernetes/staging/src/k8s.io/code-generator v1.16.2-k3s.1
	k8s.io/component-base => github.com/rancher/kubernetes/staging/src/k8s.io/component-base v1.16.2-k3s.1
	k8s.io/cri-api => github.com/rancher/kubernetes/staging/src/k8s.io/cri-api v1.16.2-k3s.1
	k8s.io/csi-translation-lib => github.com/rancher/kubernetes/staging/src/k8s.io/csi-translation-lib v1.16.2-k3s.1
	k8s.io/kube-aggregator => github.com/rancher/kubernetes/staging/src/k8s.io/kube-aggregator v1.16.2-k3s.1
	k8s.io/kube-controller-manager => github.com/rancher/kubernetes/staging/src/k8s.io/kube-controller-manager v1.16.2-k3s.1
	k8s.io/kube-proxy => github.com/rancher/kubernetes/staging/src/k8s.io/kube-proxy v1.16.2-k3s.1
	k8s.io/kube-scheduler => github.com/rancher/kubernetes/staging/src/k8s.io/kube-scheduler v1.16.2-k3s.1
	k8s.io/kubectl => github.com/rancher/kubernetes/staging/src/k8s.io/kubectl v1.16.2-k3s.1
	k8s.io/kubelet => github.com/rancher/kubernetes/staging/src/k8s.io/kubelet v1.16.2-k3s.1
	k8s.io/kubernetes => github.com/rancher/kubernetes v1.16.2-k3s.1
	k8s.io/legacy-cloud-providers => github.com/rancher/kubernetes/staging/src/k8s.io/legacy-cloud-providers v1.16.2-k3s.1
	k8s.io/metrics => github.com/rancher/kubernetes/staging/src/k8s.io/metrics v1.16.2-k3s.1
	k8s.io/node-api => github.com/rancher/kubernetes/staging/src/k8s.io/node-api v1.16.2-k3s.1
	k8s.io/sample-apiserver => github.com/rancher/kubernetes/staging/src/k8s.io/sample-apiserver v1.16.2-k3s.1
	k8s.io/sample-cli-plugin => github.com/rancher/kubernetes/staging/src/k8s.io/sample-cli-plugin v1.16.2-k3s.1
	k8s.io/sample-controller => github.com/rancher/kubernetes/staging/src/k8s.io/sample-controller v1.16.2-k3s.1
)

require (
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/ahmetb/gen-crd-api-reference-docs v0.1.5
	github.com/aokoli/goutils v1.1.0
	github.com/aws/aws-sdk-go v1.24.1
	github.com/containerd/cgroups v0.0.0-20191011165608-5fbad35c2a7e // indirect
	github.com/containerd/console v0.0.0-20181022165439-0650fd9eeb50
	github.com/containerd/containerd v1.3.0
	github.com/containerd/fifo v0.0.0-20190816180239-bda0ff6ed73c // indirect
	github.com/containerd/ttrpc v0.0.0-20191025122922-cf7f4d5f2d61 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/deislabs/smi-sdk-go v0.0.0-20190819154013-e53a9b2d8c1a
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v1.14.0-0.20190319215453-e7b5f7dbe98c
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/docker/go-units v0.4.0
	github.com/drone/envsubst v1.0.2
	github.com/envoyproxy/go-control-plane v0.8.2 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.0.6 // indirect
	github.com/fatih/color v1.7.0
	github.com/go-bindata/go-bindata v3.1.1+incompatible
	github.com/gogo/protobuf v1.3.0
	github.com/hashicorp/consul/api v1.2.0 // indirect
	github.com/jetstack/cert-manager v0.11.0
	github.com/linkerd/linkerd2 v0.0.0-20191010175117-1039d8254738
	github.com/mattn/go-shellwords v1.0.5
	github.com/moby/buildkit v0.6.2
	github.com/onsi/ginkgo v1.10.1
	github.com/pkg/browser v0.0.0-20170505125900-c90ca0c84f15
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/radovskyb/watcher v1.0.7 // indirect
	github.com/rancher/gitwatcher v0.4.4
	github.com/rancher/rdns-server v0.5.7-0.20190927164127-7128efe7d065
	github.com/rancher/wrangler v0.4.1
	github.com/rancher/wrangler-api v0.2.1-0.20191025043713-b1ca9c21825a
	github.com/sclevine/spec v1.4.0
	github.com/sirupsen/logrus v1.4.2
	github.com/solo-io/gloo v1.0.0
	github.com/solo-io/go-utils v0.10.21 // indirect
	github.com/solo-io/solo-kit v0.11.11
	github.com/stretchr/testify v1.4.0
	github.com/tektoncd/pipeline v0.8.0
	github.com/urfave/cli v1.22.2
	github.com/urfave/cli/v2 v2.1.1
	github.com/wercker/stern v1.11.0
	golang.org/x/crypto v0.0.0-20190829043050-9756ffdc2472
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	gopkg.in/inf.v0 v0.9.1
	gopkg.in/yaml.v2 v2.2.4
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.17.0
	k8s.io/apiextensions-apiserver v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/apiserver v0.0.0
	k8s.io/cli-runtime v0.0.0
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v0.4.0
	k8s.io/kubectl v0.0.0
	k8s.io/kubernetes v1.13.2
	knative.dev/pkg v0.0.0-20191024223035-2a3fc371d326
	sigs.k8s.io/yaml v1.1.0
)

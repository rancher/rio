module github.com/rancher/rio

go 1.12

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.2.0+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.1
	github.com/containerd/containerd v1.3.0-0.20190507210959-7c1e88399ec0 => github.com/containerd/containerd v1.2.1-0.20190507210959-7c1e88399ec0
	// needed for containerd
	github.com/docker/distribution => github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible
	github.com/docker/docker => github.com/docker/docker v1.4.2-0.20190319215453-e7b5f7dbe98c
	github.com/docker/docker v1.14.0-0.20190319215453-e7b5f7dbe98c => github.com/docker/docker v1.4.2-0.20190319215453-e7b5f7dbe98c
	github.com/hashicorp/vault => github.com/hashicorp/vault v1.3.2
	github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305
	github.com/jetstack/cert-manager v0.11.0 => github.com/rancher/cert-manager v0.5.1-0.20191021233300-3a070253aeda
	github.com/linkerd/linkerd2 => github.com/StrongMonkey/linkerd2 v0.5.1-0.20201008181831-07ef85ee9b6f
	github.com/pseudomuto/protoc-gen-doc => github.com/pseudomuto/protoc-gen-doc v1.0.0
	github.com/wercker/stern v1.11.0 => github.com/rancher/stern v0.0.0-20191213223518-59c2bf84f705
	golang.org/x/crypto v0.0.0-20190129210102-0709b304e793 => golang.org/x/crypto v0.0.0-20180904163835-0709b304e793
	gopkg.in/jcmturner/gokrb5.v7 => github.com/jcmturner/gokrb5 v7.3.0+incompatible
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
	k8s.io/utils => k8s.io/utils v0.0.0-20190801114015-581e00157fb1
)

require (
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/ahmetb/gen-crd-api-reference-docs v0.1.5
	github.com/aokoli/goutils v1.1.0
	github.com/aws/aws-sdk-go v1.30.16
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1 // indirect
	github.com/containerd/cgroups v0.0.0-20191011165608-5fbad35c2a7e // indirect
	github.com/containerd/console v0.0.0-20181022165439-0650fd9eeb50
	github.com/containerd/containerd v1.3.3
	github.com/containerd/fifo v0.0.0-20190816180239-bda0ff6ed73c // indirect
	github.com/containerd/ttrpc v0.0.0-20191025122922-cf7f4d5f2d61 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/deislabs/smi-sdk-go v0.1.0
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v1.14.0-0.20190319215453-e7b5f7dbe98c
	github.com/docker/go-events v0.0.0-20190806004212-e31b211e4f1c // indirect
	github.com/docker/go-units v0.4.0
	github.com/drone/envsubst v1.0.2
	github.com/fatih/color v1.9.0
	github.com/go-acme/lego v2.5.0+incompatible
	github.com/go-acme/lego/v3 v3.1.0
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/gogo/protobuf v1.3.1
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/vault/api v1.0.5-0.20200117231345-460d63e36490 // indirect
	github.com/hashicorp/vault/sdk v0.1.14-0.20200121232954-73f411823aa0 // indirect
	github.com/linkerd/linkerd2 v0.0.0-20191010175117-1039d8254738
	github.com/mattn/go-shellwords v1.0.9
	github.com/moby/buildkit v0.6.2
	github.com/onsi/ginkgo v1.11.0
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4
	github.com/pkg/errors v0.9.1
	github.com/rancher/gitwatcher v0.4.5
	github.com/rancher/norman v0.0.0-20191114233102-966e8db9e670
	github.com/rancher/rdns-server v0.5.7-0.20190927164127-7128efe7d065
	github.com/rancher/wrangler v0.2.1-0.20191205190617-661f00f286d2
	github.com/rancher/wrangler-api v0.2.1-0.20191015045805-d3635aa0853a
	github.com/sclevine/spec v1.4.0
	github.com/sirupsen/logrus v1.4.2
	github.com/solo-io/gloo v1.4.6
	github.com/solo-io/solo-kit v0.13.8
	github.com/stretchr/testify v1.5.1
	github.com/tektoncd/pipeline v0.14.3
	github.com/urfave/cli v1.22.2
	github.com/wercker/stern v1.11.0
	golang.org/x/crypto v0.0.0-20200323165209-0ec3e9974c59
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	gopkg.in/inf.v0 v0.9.1
	gopkg.in/square/go-jose.v2 v2.4.1 // indirect
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools v2.2.0+incompatible
	istio.io/api v0.0.0-20200518203817-6d29a38039bd
	istio.io/client-go v0.0.0-20200528222059-5465d5e00a32
	k8s.io/api v0.18.1
	k8s.io/apiextensions-apiserver v0.18.0
	k8s.io/apimachinery v0.18.1
	k8s.io/apiserver v0.0.0
	k8s.io/cli-runtime v0.17.3
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.17.2
	k8s.io/kubernetes v1.17.1
	sigs.k8s.io/yaml v1.2.0
)

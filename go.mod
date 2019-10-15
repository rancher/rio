module github.com/rancher/rio

go 1.12

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.0+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.1
	github.com/docker/distribution => github.com/docker/distribution v2.7.1-0.20190104202606-0ac367fd6bee+incompatible
	github.com/envoyproxy/go-control-plane => github.com/envoyproxy/go-control-plane v0.8.2
	github.com/gogo/protobuf => github.com/gogo/protobuf v1.3.0

	github.com/golang/protobuf => github.com/golang/protobuf v1.3.1
	github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305
	github.com/jetstack/cert-manager => github.com/rancher/cert-manager v0.7.0-rio.1
	github.com/knative/pkg => github.com/rancher/pkg v0.0.0-20190514055449-b30ab9de040e
	github.com/matryer/moq => github.com/rancher/moq v0.0.0-20190404221404-ee5226d43009
	golang.org/x/tools => golang.org/x/tools v0.0.0-20190411180116-681f9ce8ac52

	k8s.io/api => k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190918201827-3de75813f604
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190918202139-0b14c719ca62
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918200256-06eb1244587a
)

require (
	cloud.google.com/go v0.41.0 // indirect
	github.com/Azure/azure-sdk-for-go v31.1.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.9.1 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.6.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/validation v0.2.0 // indirect
	github.com/Masterminds/sprig v2.18.0+incompatible
	github.com/aokoli/goutils v1.1.0
	github.com/aws/aws-sdk-go v1.25.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/caddyserver/caddy v1.0.3 // indirect
	github.com/cockroachdb/datadriven v0.0.0-20190809214429-80d97fb3cbaa // indirect
	github.com/containerd/console v0.0.0-20181022165439-0650fd9eeb50
	github.com/containerd/containerd v1.3.0-0.20190507210959-7c1e88399ec0
	github.com/coredns/federation v0.0.0-20190818181423-e032b096babe // indirect
	github.com/creack/pty v1.1.7 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/deislabs/smi-sdk-go v0.0.0-20190819154013-e53a9b2d8c1a
	github.com/docker/cli v0.0.0-20190723080722-8560f9e8cdad // indirect
	github.com/docker/docker v1.14.0-0.20190319215453-e7b5f7dbe98c
	github.com/docker/go-units v0.4.0
	github.com/drone/envsubst v0.0.0-20171016184023-f4d1a8ef8670
	github.com/envoyproxy/go-control-plane v0.8.7-0.20190906190023-ba541bc36302 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.1.0 // indirect
	github.com/gdamore/tcell v0.0.0-20190319073105-ec71b09872d7 // indirect
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/gogo/googleapis v1.2.0 // indirect
	github.com/gogo/protobuf v1.3.0
	github.com/gophercloud/gophercloud v0.2.0 // indirect
	github.com/hashicorp/vault/api v1.0.4 // indirect
	github.com/jetstack/cert-manager v0.7.2
	github.com/json-iterator/go v1.1.7 // indirect
	github.com/knative/build v0.6.0
	github.com/knative/pkg v0.0.0-20190514205332-5e4512dcb2ca
	github.com/knative/serving v0.6.1
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/linkerd/linkerd2 v0.0.0-20191010175117-1039d8254738 // indirect
	github.com/markbates/inflect v1.0.4 // indirect
	github.com/mattn/go-shellwords v1.0.5
	github.com/miekg/dns v1.1.17 // indirect
	github.com/moby/buildkit v0.6.0
	github.com/onsi/ginkgo v1.8.0
	github.com/opencontainers/runc v1.0.1-0.20190307181833-2b18fe1d885e // indirect
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/radovskyb/watcher v1.0.7 // indirect
	github.com/rancher/axe v0.0.0-20190531011056-59fcf8b44147 // indirect
	github.com/rancher/gitwatcher v0.4.2-0.20191015052408-00ce0a8a61c1
	github.com/rancher/mapper v0.0.0-20190814232720-058a8b7feb99
	github.com/rancher/norman v0.0.0-20191015045353-cc004d32fcc9
	github.com/rancher/rdns-server v0.5.7-0.20190927164127-7128efe7d065
	github.com/rancher/wrangler v0.2.1-0.20191015042916-f2a6ecca4f20
	github.com/rancher/wrangler-api v0.2.1-0.20191015045805-d3635aa0853a
	github.com/rivo/tview v0.0.0-20190319111340-8d5eba0c2f51 // indirect
	github.com/rivo/uniseg v0.0.0-20190313204849-f699dde9c340 // indirect
	github.com/sclevine/spec v1.3.0
	github.com/sirupsen/logrus v1.4.2
	github.com/solo-io/gloo v0.20.3-0.20191003200350-6f6e02641501
	github.com/solo-io/go-utils v0.10.17 // indirect
	github.com/solo-io/solo-kit v0.10.24-0.20191003192541-dc479f62f67b
	github.com/stretchr/testify v1.4.0
	github.com/tektoncd/pipeline v0.4.0
	github.com/urfave/cli v1.22.1
	golang.org/x/crypto v0.0.0-20190829043050-9756ffdc2472
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/tools v0.0.0-20190907020128-2ca718005c18
	google.golang.org/api v0.10.0 // indirect
	google.golang.org/grpc v1.23.1
	gopkg.in/yaml.v2 v2.2.2
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apiextensions-apiserver v0.0.0-20190918201827-3de75813f604
	k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/cli-runtime v0.0.0-20190918202139-0b14c719ca62
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/gengo v0.0.0-20190327210449-e17681d19d3a // indirect
	k8s.io/klog v0.4.0
	k8s.io/kubernetes v1.13.4
	k8s.io/utils v0.0.0-20190506122338-8fab8cb257d5 // indirect
	sigs.k8s.io/yaml v1.1.0
)

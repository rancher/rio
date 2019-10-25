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
	github.com/jetstack/cert-manager => github.com/rancher/cert-manager v0.5.1-0.20191021233300-3a070253aeda
	github.com/linkerd/linkerd2 => github.com/StrongMonkey/linkerd2 v0.0.0-20191021165729-976fad67457a
	github.com/matryer/moq => github.com/rancher/moq v0.0.0-20190404221404-ee5226d43009
	github.com/wercker/stern => github.com/linkerd/stern v0.0.0-20190907020106-201e8ccdff9c
	golang.org/x/tools => golang.org/x/tools v0.0.0-20190411180116-681f9ce8ac52

	k8s.io/api => k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190918201827-3de75813f604
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190918202139-0b14c719ca62
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918200256-06eb1244587a
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190311093542-50b561225d70
)

require (
	contrib.go.opencensus.io/exporter/stackdriver v0.12.7 // indirect
	github.com/Azure/go-autorest/autorest v0.9.1 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.6.0 // indirect
	github.com/Masterminds/sprig v2.18.0+incompatible
	github.com/aokoli/goutils v1.1.0
	github.com/aws/aws-sdk-go v1.25.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/containerd/console v0.0.0-20181022165439-0650fd9eeb50
	github.com/containerd/containerd v1.3.0-0.20190507210959-7c1e88399ec0
	github.com/davecgh/go-spew v1.1.1
	github.com/deislabs/smi-sdk-go v0.0.0-20190819154013-e53a9b2d8c1a
	github.com/docker/cli v0.0.0-20190723080722-8560f9e8cdad // indirect
	github.com/docker/docker v1.14.0-0.20190319215453-e7b5f7dbe98c
	github.com/docker/go-units v0.4.0
	github.com/drone/envsubst v0.0.0-20171016184023-f4d1a8ef8670
	github.com/emicklei/go-restful v2.9.5+incompatible // indirect
	github.com/envoyproxy/go-control-plane v0.8.7-0.20190906190023-ba541bc36302 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.1.0 // indirect
	github.com/fatih/color v1.7.0
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/gogo/googleapis v1.2.0 // indirect
	github.com/gogo/protobuf v1.3.0
	github.com/gophercloud/gophercloud v0.2.0 // indirect
	github.com/hashicorp/go-rootcerts v1.0.1 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/jetstack/cert-manager v0.11.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/linkerd/linkerd2 v0.0.0-20191010175117-1039d8254738
	github.com/markbates/inflect v1.0.4 // indirect
	github.com/mattn/go-shellwords v1.0.5
	github.com/miekg/dns v1.1.17 // indirect
	github.com/moby/buildkit v0.6.0
	github.com/onsi/ginkgo v1.8.0
	github.com/opencontainers/runc v1.0.1-0.20190307181833-2b18fe1d885e // indirect
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/prometheus/common v0.6.0 // indirect
	github.com/prometheus/procfs v0.0.3 // indirect
	github.com/radovskyb/watcher v1.0.7 // indirect
	github.com/rancher/gitwatcher v0.4.1
	github.com/rancher/mapper v0.0.0-20190814232720-058a8b7feb99
	github.com/rancher/norman v0.0.0-20191015045353-cc004d32fcc9
	github.com/rancher/rdns-server v0.5.7-0.20190927164127-7128efe7d065
	github.com/rancher/wrangler v0.2.1-0.20191022173830-fea752b72607
	github.com/rancher/wrangler-api v0.2.1-0.20191022174038-d313951897f9
	github.com/sclevine/spec v1.3.0
	github.com/sirupsen/logrus v1.4.2
	github.com/solo-io/gloo v0.20.3-0.20191003200350-6f6e02641501
	github.com/solo-io/go-utils v0.10.17 // indirect
	github.com/solo-io/solo-kit v0.10.24-0.20191003192541-dc479f62f67b
	github.com/stretchr/testify v1.4.0
	github.com/tektoncd/pipeline v0.8.0
	github.com/urfave/cli v1.22.1
	github.com/wercker/stern v0.0.0-20171214125149-b04b5491222d
	golang.org/x/crypto v0.0.0-20190829043050-9756ffdc2472
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	google.golang.org/grpc v1.23.1 // indirect
	gopkg.in/yaml.v2 v2.2.4
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apiextensions-apiserver v0.0.0-20190918201827-3de75813f604
	k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655
	k8s.io/apiserver v0.0.0-20190918200908-1e17798da8c1
	k8s.io/cli-runtime v0.0.0-20190918202139-0b14c719ca62
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/klog v0.4.0
	k8s.io/kubernetes v1.14.3
	knative.dev/pkg v0.0.0-20191021194725-ba3f47d9e951
	sigs.k8s.io/yaml v1.1.0
)

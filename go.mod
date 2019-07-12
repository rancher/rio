module github.com/rancher/rio

go 1.12

replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.1
	github.com/jetstack/cert-manager => github.com/rancher/cert-manager v0.7.0-rio.1
	github.com/knative/pkg => github.com/rancher/pkg v0.0.0-20190514055449-b30ab9de040e
	github.com/matryer/moq => github.com/rancher/moq v0.0.0-20190404221404-ee5226d43009
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Azure/go-autorest/autorest v0.2.0 // indirect
	github.com/MakeNowJust/heredoc v0.0.0-20171113091838-e9091a26100e // indirect
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/Masterminds/sprig v2.15.0+incompatible
	github.com/Sirupsen/logrus v0.0.0-00010101000000-000000000000 // indirect
	github.com/alecthomas/chroma v0.6.3 // indirect
	github.com/aokoli/goutils v1.0.1
	github.com/cenkalti/backoff v2.1.1+incompatible // indirect
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/docker v0.7.3-0.20190618200816-384c782721c7
	github.com/docker/go-units v0.3.3
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/drone/envsubst v0.0.0-20171016184023-f4d1a8ef8670
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/envoyproxy/go-control-plane v0.7.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.0.14 // indirect
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/gdamore/tcell v0.0.0-20190319073105-ec71b09872d7
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/googleapis v1.0.0 // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/gophercloud/gophercloud v0.2.0 // indirect
	github.com/gorilla/websocket v1.2.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hashicorp/errwrap v0.0.0-20141028054710-7554cd9344ce // indirect
	github.com/hashicorp/go-multierror v0.0.0-20161216184304-ed905158d874 // indirect
	github.com/howeyc/fsnotify v0.0.0-20151003194602-f0c08ee9c607 // indirect
	github.com/huandu/xstrings v1.0.0 // indirect
	github.com/jetstack/cert-manager v0.7.2
	github.com/knative/build v0.6.0
	github.com/knative/pkg v0.0.0-20190514205332-5e4512dcb2ca
	github.com/knative/serving v0.6.1
	github.com/mattn/go-shellwords v1.0.5
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/natefinch/lumberjack v0.0.0-20170911140457-aee462912944 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/openshift/api v3.9.0+incompatible // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.0.0 // indirect
	github.com/rancher/axe v0.0.0-20190531011056-59fcf8b44147
	github.com/rancher/gitwatcher v0.3.0
	github.com/rancher/mapper v0.0.0-20190426050457-84da984f3146
	github.com/rancher/rdns-server v0.4.2
	github.com/rancher/wrangler v0.1.4
	github.com/rancher/wrangler-api v0.1.5-0.20190619170228-c3525df45215
	github.com/rivo/tview v0.0.0-20190319111340-8d5eba0c2f51
	github.com/rivo/uniseg v0.0.0-20190313204849-f699dde9c340 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5 // indirect
	github.com/stretchr/testify v1.3.0
	github.com/tektoncd/pipeline v0.4.0
	github.com/urfave/cli v1.20.1-0.20190203184040-693af58b4d51
	go.uber.org/atomic v1.3.2 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.9.0 // indirect
	golang.org/x/crypto v0.0.0-20190313024323-a1f597ede03a
	golang.org/x/net v0.0.0-20190318221613-d196dffd7c2b // indirect
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/tools v0.0.0-20190411180116-681f9ce8ac52 // indirect
	google.golang.org/appengine v1.5.0 // indirect
	google.golang.org/grpc v1.21.1 // indirect
	gopkg.in/freddierice/go-losetup.v1 v1.0.0-20170407175016-fc9adea44124 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/yaml.v2 v2.2.2
	gotest.tools v2.2.0+incompatible // indirect
	istio.io/api v0.0.0-20190408162927-e9ab8d6a54a6
	istio.io/istio v0.0.0-20190412222632-d19179769183
	k8s.io/api v0.0.0-20190606204050-af9c91bd2759
	k8s.io/apiextensions-apiserver v0.0.0-20190606210616-f848dc7be4a4
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/cli-runtime v0.0.0-20190606211212-7b5a46666fe9 // indirect
	k8s.io/client-go v11.0.1-0.20190606204521-b8faab9c5193+incompatible
	k8s.io/helm v2.14.1+incompatible // indirect
	k8s.io/klog v0.3.1
	k8s.io/kubernetes v1.14.3
	sigs.k8s.io/kustomize v2.0.3+incompatible // indirect
	sigs.k8s.io/yaml v1.1.0
)

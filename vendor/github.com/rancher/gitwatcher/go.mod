module github.com/rancher/gitwatcher

go 1.12

replace github.com/matryer/moq => github.com/rancher/moq v0.0.0-20190404221404-ee5226d43009

require (
	github.com/drone/go-scm v1.4.0
	github.com/google/uuid v1.1.1
	github.com/gorilla/mux v1.7.1
	github.com/pkg/errors v0.8.1
	github.com/rancher/wrangler v0.1.0
	github.com/rancher/wrangler-api v0.1.1
	github.com/sirupsen/logrus v1.4.1
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/urfave/cli v1.20.0
	golang.org/x/crypto v0.0.0-20190513172903-22d7a77e9e5f // indirect
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
)

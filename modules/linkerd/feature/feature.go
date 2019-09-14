package feature

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
	"github.com/linkerd/linkerd2/pkg/tls"
	v1 "github.com/rancher/rio/pkg/apis/admin.rio.cattle.io/v1"
	"github.com/rancher/rio/pkg/constants"
	"github.com/rancher/rio/pkg/constructors"
	"github.com/rancher/rio/pkg/features"
	"github.com/rancher/rio/pkg/stack"
	"github.com/rancher/rio/pkg/template/gotemplate"
	"github.com/rancher/rio/types"
	"github.com/rancher/wrangler/pkg/start"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ProxyConfig = `
- env:
{{- if .Values.CA_PEM }}
  - name: LINKERD2_PROXY_IDENTITY_TRUST_ANCHORS
    value: |
{{indent 6 .Values.CA_PEM }}
{{- end}}
  - LINKERD2_PROXY_LOG=warn,linkerd2_proxy=info
  - LINKERD2_PROXY_DESTINATION_SVC_ADDR={{ .Values.PROXY_DESTINATION }}
  - LINKERD2_PROXY_CONTROL_LISTEN_ADDR=0.0.0.0:4190
  - LINKERD2_PROXY_ADMIN_LISTEN_ADDR=0.0.0.0:4191
  - LINKERD2_PROXY_OUTBOUND_LISTEN_ADDR=127.0.0.1:4140
  - LINKERD2_PROXY_INBOUND_LISTEN_ADDR=0.0.0.0:4143
  - LINKERD2_PROXY_DESTINATION_PROFILE_SUFFIXES=svc.cluster.local.
  - LINKERD2_PROXY_INBOUND_ACCEPT_KEEPALIVE=10000ms
  - LINKERD2_PROXY_OUTBOUND_CONNECT_KEEPALIVE=10000ms
  - _pod_ns=$(self/namespace)
  - LINKERD2_PROXY_DESTINATION_CONTEXT=ns:$(_pod_ns)
  - LINKERD2_PROXY_IDENTITY_DIR=/var/run/linkerd/identity/end-entity
  - LINKERD2_PROXY_IDENTITY_TOKEN_FILE=/var/run/secrets/kubernetes.io/serviceaccount/token
  - LINKERD2_PROXY_IDENTITY_SVC_ADDR={{ .Values.IDENTITY_DESTINATION }}
  - _pod_sa=$(self/serviceAccount)
  - _l5d_ns={{ .Values.NAMESPACE }}
  - _l5d_trustdomain=cluster.local
  - LINKERD2_PROXY_IDENTITY_LOCAL_NAME=$(_pod_sa).$(_pod_ns).serviceaccount.identity.$(_l5d_ns).$(_l5d_trustdomain)
  - LINKERD2_PROXY_IDENTITY_SVC_NAME=linkerd-identity.$(_l5d_ns).serviceaccount.identity.$(_l5d_ns).$(_l5d_trustdomain)
  - LINKERD2_PROXY_DESTINATION_SVC_NAME=linkerd-controller.$(_l5d_ns).serviceaccount.identity.$(_l5d_ns).$(_l5d_trustdomain)
  - LINKERD2_PROXY_TAP_SVC_NAME=linkerd-tap.$(_l5d_ns).serviceaccount.identity.$(_l5d_ns).$(_l5d_trustdomain)
  name: linkerd-proxy
  image: gcr.io/linkerd-io/proxy:stable-2.5.0
  livenessProbe:
  httpGet:
    path: /metrics
    port: 4191
  initialDelaySeconds: 10
  readinessProbe:
  httpGet:
    path: /ready
    port: 4191
  ports:
  - 4143/http,linkerd-proxy,internal=true
  - 4191/http,linkerd-admin,internal=true
  user: 2102
  readOnly: true
  volumes:
  - linkerd-identity-end-entity:/var/run/linkerd/identity/end-entity
- args:
  - --incoming-proxy-port
  - "4143"
  - --outgoing-proxy-port
  - "4140"
  - --proxy-uid
  - "2102"
  - --inbound-ports-to-ignore
  - 4190,4191
  - --outbound-ports-to-ignore
  - "443"
  image: gcr.io/linkerd-io/proxy-init:v1.1.0
  name: linkerd-init
  cpus: "10m"
  memory: "10Mi"
  init: true
  securityContext:
    allowPrivilegeEscalation: false
    capabilities:
      add:
      - NET_ADMIN
      - NET_RAW
    privileged: false
    readOnlyRootFilesystem: true
    runAsNonRoot: false
    runAsUser: 0`
)

func Register(ctx context.Context, rContext *types.Context) error {
	linkerdInstall, err := ConfigureLinkerdInstall(rContext)
	if err != nil {
		return err
	}

	proxyInject, err := gotemplate.Apply([]byte(ProxyConfig), map[string]string{
		"CA_PEM":               string(linkerdInstall.Data["ca"]),
		"NAMESPACE":            rContext.Namespace,
		"PROXY_DESTINATION":    fmt.Sprintf("linkerd-destination.%s.svc.cluster.local:8086", rContext.Namespace),
		"IDENTITY_DESTINATION": fmt.Sprintf("linkerd-identity.%s.svc.cluster.local:8080", rContext.Namespace),
	})
	if err != nil {
		return err
	}

	proxyInjectIdentity, err := gotemplate.Apply([]byte(ProxyConfig), map[string]string{
		"CA_PEM":               string(linkerdInstall.Data["ca"]),
		"NAMESPACE":            rContext.Namespace,
		"PROXY_DESTINATION":    fmt.Sprintf("linkerd-destination.%s.svc.cluster.local:8086", rContext.Namespace),
		"IDENTITY_DESTINATION": "localhost.:8080",
	})
	if err != nil {
		return err
	}

	proxyInjectControl, err := gotemplate.Apply([]byte(ProxyConfig), map[string]string{
		"CA_PEM":               string(linkerdInstall.Data["ca"]),
		"NAMESPACE":            rContext.Namespace,
		"PROXY_DESTINATION":    "localhost.:8086",
		"IDENTITY_DESTINATION": fmt.Sprintf("linkerd-identity.%s.svc.cluster.local:8080", rContext.Namespace),
	})
	if err != nil {
		return err
	}

	apply := rContext.Apply.WithCacheTypes(rContext.Rio.Rio().V1().Service(), rContext.Core.Core().V1().ConfigMap())
	feature := &features.FeatureController{
		FeatureName: "linkerd",
		FeatureSpec: v1.FeatureSpec{
			Description: "linkerd service mesh",
			Enabled:     constants.ServiceMeshMode == constants.ServiceMeshModeLinkerd,
		},
		SystemStacks: []*stack.SystemStack{
			stack.NewSystemStack(apply, rContext.Namespace, "linkerd-crd"),
			stack.NewSystemStack(apply, rContext.Namespace, "linkerd"),
		},
		FixedAnswers: map[string]string{
			"TAG":                   constants.LinkerdVersion,
			"NAMESPACE":             rContext.Namespace,
			"UUID":                  string(linkerdInstall.Data["uuid"]),
			"CA_PEM":                string(linkerdInstall.Data["ca"]),
			"CRT_EXPIRE":            string(linkerdInstall.Data["expire"]),
			"PROXY_INJECT_CONTROL":  string(proxyInjectControl),
			"PROXY_INJECT_IDENTITY": string(proxyInjectIdentity),
			"PROXY_INJECT":          string(proxyInject),
			"CRT_PEM":               base64.StdEncoding.EncodeToString(linkerdInstall.Data["crt"]),
			"KEY_PEM":               base64.StdEncoding.EncodeToString(linkerdInstall.Data["key"]),
			"TAP_CA":                base64.StdEncoding.EncodeToString(linkerdInstall.Data["linkerd-tap-ca"]),
			"TAP_CRT":               base64.StdEncoding.EncodeToString(linkerdInstall.Data["linkerd-tap-crt"]),
			"TAP_KEY":               base64.StdEncoding.EncodeToString(linkerdInstall.Data["linkerd-tap-key"]),
			"SP_CA":                 base64.StdEncoding.EncodeToString(linkerdInstall.Data["linkerd-sp-validator-ca"]),
			"SP_CRT":                base64.StdEncoding.EncodeToString(linkerdInstall.Data["linkerd-sp-validator-crt"]),
			"SP_KEY":                base64.StdEncoding.EncodeToString(linkerdInstall.Data["linkerd-sp-validator-key"]),
			"INJECTOR_CA":           base64.StdEncoding.EncodeToString(linkerdInstall.Data["linkerd-proxy-injector-ca"]),
			"INJECTOR_CRT":          base64.StdEncoding.EncodeToString(linkerdInstall.Data["linkerd-proxy-injector-crt"]),
			"INJECTOR_KEY":          base64.StdEncoding.EncodeToString(linkerdInstall.Data["linkerd-proxy-injector-key"]),
		},

		OnStart: func(feature *v1.Feature) error {
			return start.All(ctx, 5,
				rContext.SMI)
		},
	}
	return feature.Register()
}

func ConfigureLinkerdInstall(rContext *types.Context) (*corev1.Secret, error) {
	if constants.ServiceMeshMode != constants.ServiceMeshModeLinkerd {
		return &corev1.Secret{}, nil
	}
	install, err := rContext.Core.Core().V1().Secret().Get(rContext.Namespace, "linkerd-install-secret", metav1.GetOptions{})
	if errors.IsNotFound(err) {
		linkerdInstall := constructors.NewSecret(rContext.Namespace, "linkerd-install-secret", corev1.Secret{
			StringData: map[string]string{},
		})
		root, err := tls.GenerateRootCAWithDefaults(fmt.Sprintf("identity.%s.cluster.local", rContext.Namespace))
		if err != nil {
			return install, err
		}
		u, err := uuid.NewRandom()
		if err != nil {
			return install, err
		}
		linkerdInstall.StringData["uuid"] = u.String()
		linkerdInstall.StringData["ca"] = root.Cred.Crt.EncodeCertificatePEM()
		linkerdInstall.StringData["crt"] = root.Cred.Crt.EncodeCertificatePEM()
		linkerdInstall.StringData["key"] = root.Cred.EncodePrivateKeyPEM()
		linkerdInstall.StringData["expire"] = root.Cred.Crt.Certificate.NotAfter.String()

		// linkerd-sp-validator
		services := []string{
			"linkerd-sp-validator",
			"linkerd-tap",
			"linkerd-proxy-injector",
		}
		for _, svcName := range services {
			dnsName := fmt.Sprintf("%s.%s.svc", svcName, rContext.Namespace)
			spCa, err := root.GenerateCA(dnsName, tls.Validity{}, -1)
			if err != nil {
				return install, err
			}
			linkerdInstall.StringData[fmt.Sprintf("%s-ca", svcName)] = spCa.Cred.Crt.EncodeCertificatePEM()
			linkerdInstall.StringData[fmt.Sprintf("%s-crt", svcName)] = spCa.Cred.Crt.EncodeCertificatePEM()
			linkerdInstall.StringData[fmt.Sprintf("%s-key", svcName)] = spCa.Cred.EncodePrivateKeyPEM()
		}
		return rContext.Core.Core().V1().Secret().Create(linkerdInstall)
	}
	return install, nil
}

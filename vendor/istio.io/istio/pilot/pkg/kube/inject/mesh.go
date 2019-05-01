// Copyright 2018 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package inject

const (
	sidecarTemplateDelimBegin = "[["
	sidecarTemplateDelimEnd   = "]]"

	// nolint: lll
	parameterizedTemplate = `
[[- $proxyImageKey                  := "sidecar.istio.io/proxyImage" -]]
[[- $interceptionModeKey            := "sidecar.istio.io/interceptionMode" -]]
[[- $statusPortKey                  := "status.sidecar.istio.io/port" -]]
[[- $readinessInitialDelayKey       := "readiness.status.sidecar.istio.io/initialDelaySeconds" -]]
[[- $readinessPeriodKey             := "readiness.status.sidecar.istio.io/periodSeconds" -]]
[[- $readinessFailureThresholdKey   := "readiness.status.sidecar.istio.io/failureThreshold" -]]
[[- $readinessApplicationPortsKey   := "readiness.status.sidecar.istio.io/applicationPorts" -]]
[[- $includeOutboundIPRangesKey     := "traffic.sidecar.istio.io/includeOutboundIPRanges" -]]
[[- $excludeOutboundIPRangesKey     := "traffic.sidecar.istio.io/excludeOutboundIPRanges" -]]
[[- $includeInboundPortsKey         := "traffic.sidecar.istio.io/includeInboundPorts" -]]
[[- $excludeInboundPortsKey         := "traffic.sidecar.istio.io/excludeInboundPorts" -]]
[[- $kubevirtInterfacesKey          := "traffic.sidecar.istio.io/kubevirtInterfaces" -]]
[[- $statusPortValue                := (annotation .ObjectMeta $statusPortKey {{ .StatusPort }}) -]]
[[- $readinessInitialDelayValue     := (annotation .ObjectMeta $readinessInitialDelayKey "{{ .ReadinessInitialDelaySeconds }}") -]]
[[- $readinessPeriodValue           := (annotation .ObjectMeta $readinessPeriodKey "{{ .ReadinessPeriodSeconds }}") ]]
[[- $readinessFailureThresholdValue := (annotation .ObjectMeta $readinessFailureThresholdKey {{ .ReadinessFailureThreshold }}) -]]
[[- $readinessApplicationPortsValue := (annotation .ObjectMeta $readinessApplicationPortsKey (applicationPorts .Spec.Containers)) -]]
rewriteAppHTTPProbe: {{ .RewriteAppHTTPProbe }}
initContainers:
- name: istio-init
  image: {{ .InitImage }}
  args:
  - "-p"
  - [[ .MeshConfig.ProxyListenPort ]]
  - "-u"
  - {{ .SidecarProxyUID }}
  - "-m"
  - "[[ annotation .ObjectMeta $interceptionModeKey .ProxyConfig.InterceptionMode ]]"
  - "-i"
  - "[[ annotation .ObjectMeta $includeOutboundIPRangesKey "{{ .IncludeIPRanges }}" ]]"
  - "-x"
  - "[[ annotation .ObjectMeta $excludeOutboundIPRangesKey "{{ .ExcludeIPRanges }}" ]]"
  - "-b"
  - "[[ annotation .ObjectMeta $includeInboundPortsKey (includeInboundPorts .Spec.Containers) ]]"
  - "-d"
  - "[[ excludeInboundPort $statusPortValue (annotation .ObjectMeta $excludeInboundPortsKey "{{ .ExcludeInboundPorts }}") ]]"
  {{ "[[ if (isset .ObjectMeta.Annotations \"traffic.sidecar.istio.io/kubevirtInterfaces\") -]]" }}
  - "-k"
  - "[[ annotation .ObjectMeta $kubevirtInterfacesKey "{{ .KubevirtInterfaces }}" ]]"
  {{ "[[ end -]]" }}
  {{ if eq .ImagePullPolicy "" -}}
  imagePullPolicy: IfNotPresent
  {{ else -}}
  imagePullPolicy: {{ .ImagePullPolicy }}
  {{ end -}}
  securityContext:
    capabilities:
      add:
      - NET_ADMIN
    {{ if (or (eq .DebugMode true) (eq .Privileged true)) -}}
    privileged: true
    {{ end -}}
  restartPolicy: Always
{{ if eq .EnableCoreDump true -}}
- name: enable-core-dump
  args:
  - -c
  - sysctl -w kernel.core_pattern=/var/lib/istio/core.proxy && ulimit -c unlimited
  command:
    - /bin/sh
  image: {{ .InitImage }}
  imagePullPolicy: IfNotPresent
  resources: {}
  securityContext:
    privileged: true
{{- end }}

containers:
- name: istio-proxy
  image: [[ annotation .ObjectMeta $proxyImageKey "{{ .ProxyImage }}" ]]
  args:
  - proxy
  - sidecar
  - --configPath
  - [[ .ProxyConfig.ConfigPath ]]
  - --binaryPath
  - [[ .ProxyConfig.BinaryPath ]]
  - --serviceCluster
  [[ if ne "" (index .ObjectMeta.Labels "app") -]]
  - "[[ index .ObjectMeta.Labels "app" ]].[[ valueOrDefault .DeploymentMeta.Namespace "default" ]]"
  [[ else -]]
  - "[[ valueOrDefault .DeploymentMeta.Name "istio-proxy" ]].[[ valueOrDefault .DeploymentMeta.Namespace "default" ]]"
  [[ end -]]
  - --drainDuration
  - [[ formatDuration .ProxyConfig.DrainDuration ]]
  - --parentShutdownDuration
  - [[ formatDuration .ProxyConfig.ParentShutdownDuration ]]
  - --discoveryAddress
  - [[ .ProxyConfig.DiscoveryAddress ]]
  [[ if .ProxyConfig.GetTracing.GetLightstep -]]
  - --lightstepAddress
  - [[ .ProxyConfig.GetTracing.GetLightstep.GetAddress ]]
  - --lightstepAccessToken
  - [[ .ProxyConfig.GetTracing.GetLightstep.GetAccessToken ]]
  - --lightstepSecure=[[ .ProxyConfig.GetTracing.GetLightstep.GetSecure ]]
  - --lightstepCacertPath
  - [[ .ProxyConfig.GetTracing.GetLightstep.GetCacertPath ]]
  [[ else if .ProxyConfig.GetTracing.GetZipkin -]]
  - --zipkinAddress
  - [[ .ProxyConfig.GetTracing.GetZipkin.GetAddress ]]
  [[ end -]]
  - --connectTimeout
  - [[ formatDuration .ProxyConfig.ConnectTimeout ]]
  - --statsdUdpAddress
  - [[ .ProxyConfig.StatsdUdpAddress ]]
  - --proxyAdminPort
  - [[ .ProxyConfig.ProxyAdminPort ]]
  [[ if gt .ProxyConfig.Concurrency 0 -]]
  - --concurrency
  - [[ .ProxyConfig.Concurrency ]]
  [[ end -]]
  - --controlPlaneAuthPolicy
  - [[ .ProxyConfig.ControlPlaneAuthPolicy ]]
  - --statusPort
  - [[ $statusPortValue ]]
  - --applicationPorts
  - "[[ $readinessApplicationPortsValue ]]"
  [[ if (ne $statusPortValue "0") ]]
  readinessProbe:
    httpGet:
      path: /healthz/ready
      port: [[ $statusPortValue ]]
    initialDelaySeconds: [[ $readinessInitialDelayValue ]]
    periodSeconds: [[ $readinessPeriodValue ]]
    failureThreshold: [[ $readinessFailureThresholdValue ]]
  [[ end -]]
  ports:
  - containerPort: 15090
    protocol: TCP
    name: http-envoy-prom
  env:
  - name: POD_NAME
    valueFrom:
      fieldRef:
        fieldPath: metadata.name
  - name: POD_NAMESPACE
    valueFrom:
      fieldRef:
        fieldPath: metadata.namespace
  - name: INSTANCE_IP
    valueFrom:
      fieldRef:
        fieldPath: status.podIP
  - name: ISTIO_META_POD_NAME
    valueFrom:
      fieldRef:
        fieldPath: metadata.name
  - name: ISTIO_META_INTERCEPTION_MODE
    value: [[ annotation .ObjectMeta $interceptionModeKey .ProxyConfig.InterceptionMode ]]
  {{ if eq .ImagePullPolicy "" -}}
  imagePullPolicy: IfNotPresent
  {{ else -}}
  imagePullPolicy: {{ .ImagePullPolicy }}
  {{ end -}}
  resources:
    requests:
      cpu: 10m
  securityContext:
    {{ if (or (eq .DebugMode true) (eq .Privileged true)) -}}
    privileged: true
    {{ end -}}
    {{ if eq .DebugMode true -}}
    readOnlyRootFilesystem: false
    {{ else }}
    readOnlyRootFilesystem: true
    {{ end -}}
    [[ if eq (annotation .ObjectMeta $interceptionModeKey .ProxyConfig.InterceptionMode) "TPROXY" -]]
    capabilities:
      add:
      - NET_ADMIN
    [[ end -]]
    [[ if eq (annotation .ObjectMeta $interceptionModeKey .ProxyConfig.InterceptionMode) "TPROXY" -]]
    runAsUser: 1337
    {{- if and .SDSEnabled .EnableSdsTokenMount }}
    runAsGroup: 1337
    {{ end -}}
    [[- end ]]
  volumeMounts:
  - mountPath: /etc/istio/proxy
    name: istio-envoy
  - mountPath: /etc/certs/
    name: istio-certs
    readOnly: true
{{ if eq .SDSEnabled true -}}
  - mountPath: /var/run/sds
    name: sdsudspath
{{ if eq .EnableSdsTokenMount true -}}
  - mountPath: /var/run/secrets/tokens
    name: istio-token
{{ end -}}
{{ end -}}
volumes:
{{ if eq .SDSEnabled true -}}
- name: sdsudspath
  hostPath:
    path: /var/run/sds
{{ if eq .EnableSdsTokenMount true -}}
- name: istio-token
  projected:
    sources:
    - serviceAccountToken:
      path: istio-token
      expirationSeconds: 43200
      audience: istio
{{ end -}}
{{ end -}}
- emptyDir:
    medium: Memory
  name: istio-envoy
- name: istio-certs
  secret:
    optional: true
    [[ if eq .Spec.ServiceAccountName "" -]]
    secretName: istio.default
    [[ else -]]
    secretName: [[ printf "istio.%s" .Spec.ServiceAccountName ]]
    [[ end -]]
`
)

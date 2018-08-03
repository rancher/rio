package config

var (
	parameterizedTemplate = `
initContainers:
- name: istio-init
  image: {{ .InitImage }}
  args:
  - "-p"
  - [[ .MeshConfig.ProxyListenPort ]]
  - "-u"
  - {{ .SidecarProxyUID }}
  - "-m"
  - [[ or (index .ObjectMeta.Annotations "sidecar.istio.io/interceptionMode") .ProxyConfig.InterceptionMode.String ]]
  - "-i"
  [[ if (isset .ObjectMeta.Annotations "traffic.sidecar.istio.io/includeOutboundIPRanges") -]]
  - "[[ index .ObjectMeta.Annotations "traffic.sidecar.istio.io/includeOutboundIPRanges"]]"
  [[ else -]]
  - "{{ .IncludeIPRanges }}"
  [[ end -]]
  - "-x"
  [[ if (isset .ObjectMeta.Annotations "traffic.sidecar.istio.io/excludeOutboundIPRanges") -]]
  - "[[ index .ObjectMeta.Annotations "traffic.sidecar.istio.io/excludeOutboundIPRanges" ]]"
  [[ else -]]
  - "{{ .ExcludeIPRanges }}"
  [[ end -]]
  - "-b"
  [[ if (isset .ObjectMeta.Annotations "traffic.sidecar.istio.io/includeInboundPorts") -]]
  - "[[ index .ObjectMeta.Annotations "traffic.sidecar.istio.io/includeInboundPorts" ]]"
  [[ else -]]
  - [[ range .Spec.Containers -]]
      [[ range .Ports -]]
        [[ .ContainerPort -]],
      [[ end -]]
    [[ end -]]
  [[ end ]]
  - "-d"
  [[ if (isset .ObjectMeta.Annotations "traffic.sidecar.istio.io/excludeInboundPorts") -]]
  - "[[ index .ObjectMeta.Annotations "traffic.sidecar.istio.io/excludeInboundPorts" ]]"
  [[ else -]]
  - "{{ .ExcludeInboundPorts }}"
  [[ end -]]
  {{ if eq .ImagePullPolicy "" -}}
  imagePullPolicy: IfNotPresent
  {{ else -}}
  imagePullPolicy: {{ .ImagePullPolicy }}
  {{ end -}}
  securityContext:
    capabilities:
      add:
      - NET_ADMIN
    {{ if eq .DebugMode true -}}
    privileged: true
    {{ end -}}
  restartPolicy: Always
{{ if eq .EnableCoreDump true -}}
- args:
  - -c
  - sysctl -w kernel.core_pattern=/etc/istio/proxy/core.%e.%p.%t && ulimit -c unlimited
  command:
  - /bin/sh
  image: {{ .InitImage }}
  imagePullPolicy: IfNotPresent
  name: enable-core-dump
  resources: {}
  securityContext:
    privileged: true
{{ end -}}
containers:
- name: istio-proxy
  image: [[ if (isset .ObjectMeta.Annotations "sidecar.istio.io/proxyImage") -]]
  "[[ index .ObjectMeta.Annotations "sidecar.istio.io/proxyImage" ]]"
  [[ else -]]
  {{ .ProxyImage }}
  [[ end -]]
  args:
  - proxy
  - sidecar
  - --configPath
  - [[ .ProxyConfig.ConfigPath ]]
  - --binaryPath
  - [[ .ProxyConfig.BinaryPath ]]
  - --serviceCluster
  [[ if ne "" (index .ObjectMeta.Labels "app") -]]
  - [[ index .ObjectMeta.Labels "app" ]]
  [[ else -]]
  - "istio-proxy"
  [[ end -]]
  - --drainDuration
  - [[ formatDuration .ProxyConfig.DrainDuration ]]
  - --parentShutdownDuration
  - [[ formatDuration .ProxyConfig.ParentShutdownDuration ]]
  - --discoveryAddress
  - [[ .ProxyConfig.DiscoveryAddress ]]
  - --discoveryRefreshDelay
  - [[ formatDuration .ProxyConfig.DiscoveryRefreshDelay ]]
  - --zipkinAddress
  - [[ .ProxyConfig.ZipkinAddress ]]
  - --connectTimeout
  - [[ formatDuration .ProxyConfig.ConnectTimeout ]]
  - --statsdUdpAddress
  - [[ .ProxyConfig.StatsdUdpAddress ]]
  - --proxyAdminPort
  - [[ .ProxyConfig.ProxyAdminPort ]]
  - --controlPlaneAuthPolicy
  - [[ .ProxyConfig.ControlPlaneAuthPolicy ]]
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
    value: [[ or (index .ObjectMeta.Annotations "sidecar.istio.io/interceptionMode") .ProxyConfig.InterceptionMode.String ]]
  {{ if eq .ImagePullPolicy "" -}}
  imagePullPolicy: IfNotPresent
  {{ else -}}
  imagePullPolicy: {{ .ImagePullPolicy }}
  {{ end -}}
  #resources:
  #  requests:
  #    cpu: 100m
  #    memory: 128Mi
  securityContext:
    {{ if eq .DebugMode true -}}
    privileged: true
    readOnlyRootFilesystem: false
    {{ else -}}
    privileged: false
    readOnlyRootFilesystem: true
    [[ if eq (or (index .ObjectMeta.Annotations "sidecar.istio.io/interceptionMode") .ProxyConfig.InterceptionMode.String) "TPROXY" -]]
    capabilities:
      add:
      - NET_ADMIN
    [[ end -]]
    {{ end -}}
    [[ if ne (or (index .ObjectMeta.Annotations "sidecar.istio.io/interceptionMode") .ProxyConfig.InterceptionMode.String) "TPROXY" -]]
    runAsUser: 1337
    [[ end -]]
  restartPolicy: Always
  volumeMounts:
  - mountPath: /etc/istio/proxy
    name: istio-envoy
  - mountPath: /etc/certs/
    name: istio-certs
    readOnly: true
volumes:
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

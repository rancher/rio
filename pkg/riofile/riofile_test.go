package riofile

import (
	"fmt"
	"testing"

	"github.com/rancher/rio/pkg/template"
	"github.com/rancher/wrangler/pkg/yaml"
)

const (
	file = `
configs:
  logging:
    content: |-
      loglevel.controller: info
        loglevel.creds-init: info
        loglevel.git-init: info
        loglevel.webhook: info
        zap-logger-config: |
          {
            "level": "info",
            "development": false,
            "sampling": {
              "initial": 100,
              "thereafter": 100
            },
            "outputPaths": ["stdout"],
            "errorOutputPaths": ["stderr"],
            "encoding": "json",
            "encoderConfig": {
              "timeKey": "",
              "levelKey": "level",
              "nameKey": "logger",
              "callerKey": "caller",
              "messageKey": "msg",
              "stacktraceKey": "stacktrace",
              "lineEnding": "",
              "levelEncoder": "",
              "timeEncoder": "",
              "durationEncoder": "",
              "callerEncoder": ""
            }
          }
services:
  build-controller:
    global_permissions:
    - '* pods'
    - '* namespaces'
    - '* secrets'
    - '* events'
    - '* serviceaccounts'
    - '* configmaps'
    - '* extensions/deployments'
    - 'create,get,list,watch,patch,update,delete build.knative.dev/builds'
    - 'create,get,list,watch,patch,update,delete build.knative.dev/builds/status'
    - 'create,get,list,watch,patch,update,delete build.knative.dev/buildtemplates'
    - 'create,get,list,watch,patch,update,delete build.knative.dev/clusterbuildtemplates'
    - '* caching.internal.knative.dev/images'
    - '* apiextensions.k8s.io/customresourcedefinitions'
    image: gcr.io/knative-releases/github.com/knative/build/cmd/controller@sha256:6c9133810e75c057e6084f5dc65b6c55cb98e42692f45241f8d0023050f27ba9
    configs:
    - logging:/etc/config-logging
    environment:
    - SYSTEM_NAMESPACE=${NAMESPACE}
    command:
    - -logtostderr
    - -stderrthreshold
    - INFO
    - -creds-image
    - gcr.io/knative-releases/github.com/knative/build/cmd/creds-init@sha256:22b3a971c3d1d5529ca16f6b6d168ba03c1f3bcb0744271ff8882374fd3b6fdb
    - -git-image
    - gcr.io/knative-releases/github.com/knative/build/cmd/git-init@sha256:e6ffa2a922cdea55d51d8648b5b07435d5598ebb6789849c41802de63e7324a9
    - -nop-image
    - gcr.io/knative-releases/github.com/knative/build/cmd/nop@sha256:915db860d1bf101322f35b06e963a1dcc00e9c1beeecfaaef650db4e45364e61



kubernetes:
  namespaced_custom_resource_definitions:
  - BuildTemplate.build.knative.dev/v1alpha1
  - Image.caching.internal.knative.dev/v1alpha1
  custom_resource_definitions:
  - ClusterBuildTemplate.build.knative.dev/v1alpha1
  manifest: |-
    apiVersion: caching.internal.knative.dev/v1alpha1
    kind: Image
    metadata:
      name: creds-init
      namespace: knative-build
    spec:
      image: gcr.io/knative-releases/github.com/knative/build/cmd/creds-init@sha256:22b3a971c3d1d5529ca16f6b6d168ba03c1f3bcb0744271ff8882374fd3b6fdb
    ---
    apiVersion: caching.internal.knative.dev/v1alpha1
    kind: Image
    metadata:
      name: git-init
      namespace: knative-build
    spec:
      image: gcr.io/knative-releases/github.com/knative/build/cmd/git-init@sha256:e6ffa2a922cdea55d51d8648b5b07435d5598ebb6789849c41802de63e7324a9
    ---
    apiVersion: caching.internal.knative.dev/v1alpha1
    kind: Image
    metadata:
      name: gcs-fetcher
      namespace: knative-build
    spec:
      image: gcr.io/cloud-builders/gcs-fetcher
    ---
    apiVersion: caching.internal.knative.dev/v1alpha1
    kind: Image
    metadata:
      name: nop
      namespace: knative-build
    spec:
      image: gcr.io/knative-releases/github.com/knative/build/cmd/nop@sha256:915db860d1bf101322f35b06e963a1dcc00e9c1beeecfaaef650db4e45364e61
    ---
    apiVersion: build.knative.dev/v1alpha1
    kind: ClusterBuildTemplate
    metadata:
      name: buildkit
    spec:
      parameters:
      - name: IMAGE
        description: Where to publish the resulting image
      - name: DOCKERFILE
        description: The name of the Dockerfile
        default: "Dockerfile"
      - name: PUSH
        description: Whether push or not
        default: "true"
      - name: DIRECTORY
        description: The directory containing the app
        default: "/workspace"
      - name: BUILDKIT_CLIENT_IMAGE
        description: The name of the BuildKit client (buildctl) image
        default: "moby/buildkit:v0.3.1-rootless@sha256:2407cc7f24e154a7b699979c7ced886805cac67920169dcebcca9166493ee2b6"
      - name: BUILDKIT_DAEMON_ADDRESS
        description: The address of the BuildKit daemon (buildkitd) service
        default: "tcp://buildkitd:1234"
      steps:
      - name: build-and-push
        image: $${BUILDKIT_CLIENT_IMAGE}
        workingDir: $${DIRECTORY}
        command: ["buildctl", "--addr=$${BUILDKIT_DAEMON_ADDRESS}", "build",
                  "--progress=plain",
                  "--frontend=dockerfile.v0",
                  "--frontend-opt", "filename=$${DOCKERFILE}",
                  "--local", "context=.", "--local", "dockerfile=.",
                  "--exporter=image", "--exporter-opt", "name=$${IMAGE}", "--exporter-opt", "push=$${PUSH}"]
`
)

func TestParse(t *testing.T) {
	t.Skip("Fix post 0.6 RC")
	f, err := Parse([]byte(file), template.AnswersFromMap(nil))
	if err != nil {
		t.Fatal(err)
	}

	if len(f.Services) != 1 {
		t.Fatal("expected 1")
	}

	obj := f.Objects()
	yamlBytes, err := yaml.ToBytes(obj)
	if err != nil {
		t.Fatal(err)
	}

	if len(f.Kubernetes) == 0 {
		t.Fatal("not enough objects")
	}

	if len(f.Objects()) != 3 {
		t.Fatal("expected three")
	}

	fmt.Println(string(yamlBytes))

}

#!/bin/bash

cat > tekton-stack.yaml << EOF
kubernetes:
  manifest: |
$(curl -sL https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml | sed 's/^/    /g')

template:
  goTemplate: false
EOF
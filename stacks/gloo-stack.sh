#!/bin/bash

cat > gloo-stack.yaml << EOF
kubernetes:
  manifest: |
$(glooctl install gateway -n '${NAMESPACE}' --values ./gloo-values.yaml --dry-run | sed 's/^/    /g')

template:
  goTemplate: false
EOF

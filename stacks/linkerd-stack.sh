#!/bin/bash

cat > linkerd-stack.yaml << EOF
kubernetes:
  manifest: |
$(linkerd install --ignore-cluster | sed 's/^/    /g')

routers:
  web:
    routes:
    - to:
      - app: linkerd-web
        port: 8084

template:
  goTemplate: false
EOF
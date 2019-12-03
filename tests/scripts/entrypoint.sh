#!/bin/bash

if [ "${CLUSTER}" == "k3s" ]; then
  source ./tests/scripts/install_k3s.sh
elif [ "${CLUSTER}" == "gke" ]; then
  source ./tests/scripts/install_gke.sh
elif [ "${CLUSTER}" == "rke" ]; then
  source ./tests/scripts/install_rke.sh
else
  echo "Using given cluster with given kubeconfig..."
fi

# Get rio binary
curl -sfL https://get.rio.io | sh - > /dev/null 2>&1

# Install rio if it isn't already installed
if ! [ "$(rio info | grep "Cluster Domain IPs")" ] ; then rio install ; fi

if [ "${CLUSTER}" == "k3s" ]; then kubectl delete svc traefik -n kube-system ; fi

exec "$@"

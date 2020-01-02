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
if [ "${RIO_VERSION}" == "master" ]; then
  echo "Installing rio from source"
  ./scripts/build && docker build -f package/Dockerfile -t ${REPO}/rio-controller:${TAG} -q --force-rm .
  go build -ldflags " -X github.com/rancher/rio/pkg/constants.ControllerImage=${REPO}/rio-controller -X github.com/rancher/rio/pkg/constants.ControllerImageTag=${TAG}" -i -o bin/rio cli/main.go
  cp ./bin/rio /usr/local/bin/
else
  curl -sfL https://get.rio.io | sh - > /dev/null 2>&1
fi

# Install rio if it isn't already installed
if ! [ "$(rio info | grep "Cluster Domain IPs")" ] ; then rio install --no-email ; fi

if [ "${CLUSTER}" == "k3s" ]; then kubectl delete svc traefik -n kube-system ; fi

rio info
rio build-history

exec "$@"

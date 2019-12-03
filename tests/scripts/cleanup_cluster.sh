#!/bin/bash

echo "Cleaning up resources from $CLUSTER cluster..."

if [ "${CLUSTER}" == "k3s" ]; then
  # Delete droplets
  for ((i=0; i<=$NUM_WORKERS; i++))
  do
    curl -s -X DELETE -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" "https://api.digitalocean.com/v2/droplets/$(eval echo \${NODEID_$i})"
  done
  # Delete ssh key
  curl -s -X DELETE -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" "https://api.digitalocean.com/v2/account/keys/$(echo $SSH_KEY | jq -r .ssh_key.id)"
elif [ "${CLUSTER}" == "gke" ]; then
  echo "cleanup gke..."
elif [ "${CLUSTER}" == "rke" ]; then
  echo "cleanup rke..."
fi

echo "$CLUSTER cluster resources deleted."

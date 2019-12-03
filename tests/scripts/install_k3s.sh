#!/bin/bash

echo "Configuring k3s cluster on Digital Ocean..."

# Create new ssh-key
create_ssh_key() {
  SSH_KEY_NAME="rio-test-"$(cat /dev/random | LC_CTYPE=C tr -dc "[:alnum:]" | head -c 5)
  ssh-keygen -t rsa -N "" -f $SSH_KEY_NAME.key -q
  SSH_KEY=$(curl -s -X POST -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d"{\"name\":\"$SSH_KEY_NAME\",\"public_key\":\"$(cat $SSH_KEY_NAME.key.pub)\"}" "https://api.digitalocean.com/v2/account/keys")
  export SSH_KEY
}

create_nodes() {
  # Create nodes in Digital Ocean
  DO_RESULT_NODES=()
  for ((i=0; i<=$1; i++))
  do
    DO_RESULT_NODES[$i]=$(curl -s -X POST "https://api.digitalocean.com/v2/droplets" -d"{\"names\":[\"rio-automated-k3s-node-$((i + 1))\"],\"region\":\"sfo2\",\"size\":\"s-2vcpu-4gb\",\"image\":\"ubuntu-18-04-x64\",\"ssh_keys\":[$(echo $SSH_KEY | jq -r .ssh_key.id)]}" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json")
  done

  # Wait until nodes statuses are active
  for ((i=0; i<=$1; i++))
  do
    DO_RESULT_NODES[$i]=$(curl -s -X GET -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" "https://api.digitalocean.com/v2/droplets/$(echo ${DO_RESULT_NODES[$i]} | jq .droplets[0].id)")
    until [ $(echo ${DO_RESULT_NODES[$i]} | jq -r .droplet.status) == "active" ]
    do
      sleep 5
      DO_RESULT_NODES[$i]=$(curl -s -X GET -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" "https://api.digitalocean.com/v2/droplets/$(echo ${DO_RESULT_NODES[$i]} | jq .droplet.id)")
    done
  done

  for ((i=0; i<=$1; i++))
  do
    export NODEID_$i=$(echo ${DO_RESULT_NODES[$i]} | jq .droplet.id)
  done
  sleep 30
}

install_k3s_on_master() {
  ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i $SSH_KEY_NAME.key root@$(echo ${DO_RESULT_NODES[0]} | jq -r .droplet.networks.v4[0].ip_address) /bin/bash <<- EOF
    curl -sfL https://get.k3s.io | sh -
EOF
}

add_k3s_agents() {
  for ((i=1; i<=$1; i++))
  do
    ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i $SSH_KEY_NAME.key root@$(echo ${DO_RESULT_NODES[$i]} | jq -r .droplet.networks.v4[0].ip_address) /bin/bash <<- EOF
      curl -sfL https://get.k3s.io | K3S_URL=https://$(echo ${DO_RESULT_NODES[$i]} | jq -r .droplet.networks.v4[0].ip_address):6443 K3S_TOKEN="$(echo $NODE_TOKEN)" sh -
EOF
  done
}

export NUM_WORKERS=${WORKERS-2}
create_ssh_key
create_nodes NUM_WORKERS
install_k3s_on_master

# Extract NODE_TOKEN and KUBECONFIG from master node
NODE_TOKEN=$(ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i $SSH_KEY_NAME.key root@$(echo ${DO_RESULT_NODES[0]} | jq -r .droplet.networks.v4[0].ip_address) "cat /var/lib/rancher/k3s/server/node-token")
LOCAL_CONFIG=$(ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i $SSH_KEY_NAME.key root@$(echo ${DO_RESULT_NODES[0]} | jq -r .droplet.networks.v4[0].ip_address) "kubectl config view --flatten -o json")
echo "${LOCAL_CONFIG//127.0.0.1/$(echo ${DO_RESULT_NODES[0]} | jq -r .droplet.networks.v4[0].ip_address)}" > .kube/config

add_k3s_agents NUM_WORKERS
echo "k3s cluster successfully configured with $((NUM_WORKERS + 1)) nodes."

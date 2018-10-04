#!/usr/bin/env bash
set -x

setup() {
    export nfs_server_stack=nfs-server-${RANDOM}
    export nfs_client_stack=nfs-client-${RANDOM}
    server=$(rio up $nfs_server_stack ./tests/nfs-stack/nfs-server-stack.yaml)
    export server_state=""
    while [[ $server_state != "active" ]]; do
        server_state=$(rio inspect ${nfs_server_stack}/nfs-server | jq -r '.state')
        sleep 1
    done
    export ip=""
    while ! [[ $ip =~ ^10.42.* ]]; do
        ip=$(rio kubectl get po --all-namespaces -o wide | grep nfs-server | awk '{print $(NF-1)}')
        sleep 1
    done
    printf "NFS_SERVER_HOSTNAME: $ip\nNFS_SERVER_EXPORT_PATH: /" > answers.yaml
    rio up --answers answers.yaml $nfs_client_stack ./tests/nfs-stack/nfs-client-stack.yaml
    export client_state=""
    while [[ $client_state != "active" ]]; do
        client_state=$(rio inspect ${nfs_client_stack}/nfs-provisioner | jq -r '.state')
        sleep 1
    done
    rm answers.yaml
}

setup

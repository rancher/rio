#!/usr/bin/env bash
set -x

setup() {
    nfs_server_stack=nfs-server-${RANDOM}
    nfs_client_stack=nfs-client-${RANDOM}
    rio up $nfs_server_stack ./tests/nfs-stack/nfs-server-stack.yaml
    rio wait ${nfs_server_stack}/nfs-server
    ip=$(rio kubectl get po --all-namespaces -o wide | grep nfs-server | awk '{print $(NF-1)}')
    printf "NFS_SERVER_HOSTNAME: $ip\nNFS_SERVER_EXPORT_PATH: /" > answers.yaml
    rio up --answers answers.yaml $nfs_client_stack ./tests/nfs-stack/nfs-client-stack.yaml
    rio wait ${nfs_client_stack}/nfs-provisioner
    rm answers.yaml
}

setup

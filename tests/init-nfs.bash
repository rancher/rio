#!/usr/bin/env bash
set -x

setup() {
    nfs_server_stack=nfs-server-${RANDOM}
    rio up $nfs_server_stack ./tests/nfs-stack/nfs-server-stack.yaml
    rio wait ${nfs_server_stack}/nfs-server
    ip=$(rio kubectl get po --all-namespaces -o wide | grep nfs-server | awk '{print $(NF-2)}')
    printf "NFS_SERVER_HOSTNAME: $ip\nNFS_SERVER_EXPORT_PATH: /" > answers.yaml
    rio feature enable -a answers.yaml nfs
    rm answers.yaml
}

setup

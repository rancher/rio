## Setup ##

setup() {
    export nfs_volume_name=nfs-test-${RANDOM}
    rio volume create -d nfs $nfs_volume_name 1
}

teardown() {
    rio rm $nfs_volume_name
}

load volume-common

@test "nfs volume is bound" {
    if [[ $RUN_NFS_TEST != "true" ]]; then
        skip "RUN_NFS_TEST IS NOT ENABLED"
    fi
    wait_for_state $nfs_volume_name "bound"
}

@test "bound nfs volume to a workload" {
    if [[ $RUN_NFS_TEST != "true" ]]; then
        skip "RUN_NFS_TEST IS NOT ENABLED"
    fi
    wait_for_state $nfs_volume_name "bound"
    workload1=$(rio run -v $nfs_volume_name:/data nginx | cut -d':' -f2)
    wait_for_state $workload1 "active"
    rio exec $workload1 touch /data/helloworld
    workload2=$(rio run -v $nfs_volume_name:/data nginx | cut -d':' -f2)
    wait_for_state $workload2 "active"
    output=$(rio exec $workload2 ls /data)
    [ $output == "helloworld" ]
    rio rm $workload1 $workload2
}
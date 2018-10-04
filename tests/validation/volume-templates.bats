load volume-common

setup() {
    export stack_name=template-${RANDOM}
}

teardown() {
    rio rm $stack_name
}

@test "volume templates" {
    if [[ $RUN_NFS_TEST != "true" ]]; then
        skip "RUN_NFS_TEST IS NOT ENABLED"
    fi
    rio up ${stack_name} ./tests/nfs-stack/volume-template-stack.yaml
    wait_for_ip test1-0
    wait_for_ip test2-0
    template=$(rio inspect ${stack_name}/data --format json | jq '.template')
    [ $template == "true" ]
    wait_for_state ${stack_name}/data-test1-0 "bound"
    wait_for_state ${stack_name}/data-test2-0 "bound"
    rio exec ${stack_name}/test1 touch /persistentvolumes/helloworld
    rio exec ${stack_name}/test2 touch /persistentvolumes/helloworld
    rio run -v data-test1-0:/data --name ${stack_name}/inspect-v1 nginx
    rio run -v data-test2-0:/data --name ${stack_name}/inspect-v2 nginx
    wait_for_ip inspect-v1
    wait_for_ip inspect-v2
    output1=$(rio exec ${stack_name}/inspect-v1 ls /data)
    output2=$(rio exec ${stack_name}/inspect-v2 ls /data)
    [ $output1 == helloworld ]
    [ $output2 == helloworld ]
}
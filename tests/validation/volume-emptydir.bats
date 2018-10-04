load volume-common

@test "anonymous volume" {
    workload=$(rio run -v /data nginx | cut -d':' -f2)
    wait_for_state $workload "active"
    wait_for_ip $workload
    pod_name=$(rio kubectl get pod --all-namespaces | grep $workload | awk '{print $2}')
    ns=$(rio kubectl get pod --all-namespaces | grep $workload | awk '{print $1}' )
    volume_name=$(rio kubectl get po $pod_name -n $ns -o json | jq '.spec.containers[0].volumeMounts[0].name' | tr -d '"')
    mount_point=$(rio kubectl get po $pod_name -n $ns -o json | jq '.spec.containers[0].volumeMounts[0].mountPath' | tr -d '"')
    mount_type=$(rio kubectl get po $pod_name -n $ns -o json | jq '.spec.volumes[0] | keys[0]' | tr -d '"')
    [ $volume_name == "anon-data" ]
    [ $mount_point == "/data" ]
    [ $mount_type == "emptyDir" ]
    rio rm $workload
}

@test "named volume" {
    workload=$(rio run -v test:/data nginx | cut -d':' -f2)
    wait_for_state $workload "active"
    wait_for_ip $workload
    pod_name=$(rio kubectl get pod --all-namespaces | grep $workload | awk '{print $2}')
    ns=$(rio kubectl get pod --all-namespaces | grep $workload | awk '{print $1}' )
    rio kubectl get po $pod_name -n $ns -o yaml
    volume_name=$(rio kubectl get po $pod_name -n $ns -o json | jq '.spec.containers[0].volumeMounts[0].name' | tr -d '"')
    mount_point=$(rio kubectl get po $pod_name -n $ns -o json | jq '.spec.containers[0].volumeMounts[0].mountPath' | tr -d '"')
    mount_type=$(rio kubectl get po $pod_name -n $ns -o json | jq '.spec.volumes[0] | keys[0]' | tr -d '"')
    [ $volume_name == "test" ]
    [ $mount_point == "/data" ]
    [ $mount_type == "emptyDir" ]
    rio rm $workload
}
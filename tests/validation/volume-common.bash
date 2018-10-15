wait_for_state() {
    for i in {1..30}
    do
        sleep 2
        if [[ $(rio inspect $1 | jq -r '.state') == $2 ]]; then
            pass="true"
            break
        fi
    done
    [ $pass == "true" ]
}

wait_for_ip() {
    export ip=""
    for i in {1..30}
    do
        if ! [[ $ip =~ ^10.42.* ]]; then
            ip=$(rio kubectl get po --all-namespaces -o wide | grep $1 | awk '{print $(NF-2)}')
            sleep 2
        fi
    done
}
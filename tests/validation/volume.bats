## Setup ##

setup() {
  export vol=tvol${RANDOM}
  rio volume create ${vol} 10
}

teardown () {
  rio volume rm ${vol}
}

## Validation tests ##
@test "volume - volume listing & size" {
  rio volume
  [ "$(rio inspect --format '{{.name}}' ${vol})" == ${vol} ]
  [ "$(rio inspect --format '{{.sizeInGb}}' ${vol})" == "10" ]
  nsp="$(rio inspect --format '{{.id}}' ${vol} | cut -f1 -d:)"
  sleep 10
  [ "$(rio kubectl get -n ${nsp} -o=json pvc/${vol} | jq -r .metadata.name)" == "${vol}" ]
  [ "$(rio kubectl get -n ${nsp} -o=json pvc/${vol} | jq -r .spec.resources.requests.storage)" == "10Gi" ]

}

@test "k8s volume - volume size is 10Gi" {
    nsp="$(rio inspect --format '{{.id}}' ${vol} | cut -f1 -d:)"
    rio volume
    sleep 10
    [ "$(rio kubectl get -n ${nsp} -o=json pvc/${vol} | jq -r .spec.resources.requests.storage)" == "10Gi" ]
}

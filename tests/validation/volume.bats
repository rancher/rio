## Setup ##

setup() {
  export vol=tvol${RANDOM}
  rio volume create ${vol} 10
}

teardown () {
  rio volume rm ${vol}
}

## Validation tests ##
@test "rio volume - volume is listing" {
  rio volume
  [ "$(rio inspect --format '{{.name}}' ${vol})" == ${vol} ]
}

@test "rio volume - volume is bound" {
  skip
  rio volume
  sleep 10
  [ "$(rio inspect --format '{{.state}}' ${vol})" == "bound" ]
}

@test "rio volume - volume size is 10" {
  rio volume
  [ "$(rio inspect --format '{{.sizeInGb}}' ${vol})" == "10" ]
}

@test "k8s volume - volume is listing" {
    nsp="$(rio inspect --format '{{.id}}' ${vol} | cut -f1 -d:)"
    rio volume
    sleep 10
    [ "$(rio kubectl get -n ${nsp} -o=json pvc/${vol} | jq -r .metadata.name)" == "${vol}" ]
}

@test "k8s volume - volume size is 10Gi" {
    nsp="$(rio inspect --format '{{.id}}' ${vol} | cut -f1 -d:)"
    rio volume
    sleep 10
    [ "$(rio kubectl get -n ${nsp} -o=json pvc/${vol} | jq -r .spec.resources.requests.storage)" == "10Gi" ]

}
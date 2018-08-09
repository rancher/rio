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
  [[ "$(rio inspect --format '{{.name}}' ${vol})" == ${vol} ]]
}

@test "rio volume - volume is bound" {
  rio volume
  sleep 10
  [[ "$(rio inspect --format '{{.state}}' ${vol})" == "bound" ]]
}

@test "rio volume - volume size is 10" {
  rio volume
  [[ "$(rio inspect --format '{{.sizeInGb}}' ${vol})" == "10" ]]
}

@test "rio volume - volume exist in k8s" {
    nsp="$(rio inspect --format '{{.id}}' ${vol} | cut -f1 -d:)"
    rio volume
    sleep 10
    [ "$(rio kubectl get -n ${nsp} -o=json pvc/${vol} | jq -r .metadata.name)" == "${vol}" ]
}

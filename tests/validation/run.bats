## Setup ##

setup() {
  export srv=tsrv${RANDOM}
  rio run -n ${srv} nginx
  rio wait ${srv}
}

teardown () {
  rio rm ${srv}
}

## Validation tests ##
@test "rio run - service exist" {
  rio ps
  [[ "$(rio inspect --format '{{.name}}' ${srv})" == ${srv} ]] || false
}

@test "rio run - service state active" {
  rio ps
  [ "$(rio inspect --format '{{.state}}' ${srv})" == "active" ]
}

@test "rio run - service scale = 1" {
  rio ps
  [ "$(rio inspect --format '{{.scale}}' ${srv})" == "1" ]
}

@test "rio run - service is active in kubernetes" {
    nsp="$(rio inspect --format '{{.id}}' ${srv} | cut -f1 -d:)"
    [ "$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r .status.replicas)" == "1" ]
}

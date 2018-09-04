## Setup ##

setup() {
  export srv=tsrv${RANDOM}
  rio run -n ${srv} nginx
  rio scale ${srv}=3
  rio wait ${srv}
}

teardown () {
  rio rm ${srv}
}

## Validation tests ##

@test "rio scale - service is listing" {
  rio ps
  [ "$(rio inspect --format '{{.name}}' ${srv})" == ${srv} ]
}

@test "rio scale - service state active" {
  rio ps
  [ "$(rio inspect --format '{{.state}}' ${srv})" == "active" ]
}

@test "rio scale - service scale = 3" {
  rio ps
  [ "$(rio inspect --format '{{.scale}}' ${srv})" == "3" ]
}

@test "rio scale - kubectl replicas = 3" {
    nsp="$(rio inspect --format '{{.id}}' ${srv} | cut -f1 -d:)"
    [ "$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r .status.replicas)" == "3" ]
}

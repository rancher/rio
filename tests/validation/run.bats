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

@test "run - service state active" {
  rio ps
  [ "$(rio inspect --format '{{.state}}' ${srv})" == "active" ]
  nsp="$(rio inspect --format '{{.id}}' ${srv} | cut -f1 -d:)"
  [ "$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r .status.replicas)" == "1" ]

}

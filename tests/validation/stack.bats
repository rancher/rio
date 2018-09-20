## Setup ##

setup() {
  export stack=stk${RANDOM} 
  export srv=tsrv${RANDOM}
  rio run -n ${stack}/${srv} nginx
  rio wait ${stack}/${srv}
}

teardown () {
  rio rm ${stack}/${srv}
  rio stack rm ${stack}
}

## Validation tests ##
@test "stack - stack exist and listing" {
  rio ps -q ${stack}
  [ "$(rio inspect --format '{{.name}}' ${stack})" == ${stack} ]
  [ "$(rio inspect --format '{{.state}}' ${stack})" == "active" ]
  [ "$(rio inspect --format '{{.stackId}}' ${stack}/${srv} | cut -f2 -d:)" == "${stack}" ]
  nsp="$(rio inspect --format '{{.id}}' ${stack}/${srv} | cut -f1 -d:)"
  [ "$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r .status.replicas)" == "1" ]

}


@test "stack - service added to existing stack" {
  rio ps -q ${stack}
  srv2=tsrv${RANDOM}
  rio run -n ${stack}/${srv2} nginx
  rio wait ${stack}/${srv2}
  [ "$(rio inspect --format '{{.stackId}}' ${stack}/${srv2}  | cut -f2 -d:)" == "${stack}" ]
  nsp="$(rio inspect --format '{{.id}}' ${stack}/${srv2} | cut -f1 -d:)"
  [ "$(rio kubectl get -n ${nsp} -o=json deploy/${srv2} | jq -r .status.replicas)" == "1" ]

}

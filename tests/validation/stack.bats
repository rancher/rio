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
@test "rio stack - stack is listing" {
  rio ps -q ${stack}
  [ "$(rio inspect --format '{{.name}}' ${stack})" == ${stack} ]
}

@test "rio stack - stack is active" {
  rio ps -q ${stack}
  [ "$(rio inspect --format '{{.state}}' ${stack})" == "active" ]
}

@test "rio stack - serivce was added to stack" {
  rio ps -q ${stack}
  [ "$(rio inspect --format '{{.stackId}}' ${stack}/${srv} | cut -f2 -d:)" == "${stack}" ]
}

@test "rio stack - service added to existing stack" {
  rio ps -q ${stack}
  srv2=tsrv${RANDOM}
  rio run -n ${stack}/${srv2} nginx
  rio wait ${stack}/${srv2}
  [ "$(rio inspect --format '{{.stackId}}' ${stack}/${srv2}  | cut -f2 -d:)" == "${stack}" ]
}

@test "k8s stack - service exist" {
    nsp="$(rio inspect --format '{{.id}}' ${stack}/${srv} | cut -f1 -d:)"
    [ "$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r .status.replicas)" == "1" ]
}

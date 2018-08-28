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
@test "rio export - stack and service exist {
  rio ps -q ${stack}
  [[ "$(rio inspect --format '{{.name}}' ${stack})" == ${stack} ]]
  [[ "$(rio inspect --format '{{.name}}' ${stack}/${srv})" == ${srv} ]]
}

@test "rio export - service info exporting" {
  rio export -o json -t service ${stack}/${srv}
  [[ "$(rio export -o json -t service ${stack}/${srv} | jq -r .image)" == "nginx" ]]
  [[ "$(rio export -o json -t service ${stack}/${srv} | jq -r .name)" == "${srv}" ]]
  [[ "$(rio export -o json -t service ${stack}/${srv} | jq .scale)" == "1" ]]
}

@test "rio export - stack info exporting" {
  rio export ${stack}
  [[ "$(rio export -o json ${stack} | jq -r .services.${srv}.image)" == "nginx" ]]
  [[ "$(rio export -o json ${stack} | jq .services.${srv}.scale)" == "1" ]]
}


## Setup ##

setup() {
  export srv=tsrv${RANDOM}
  rio create -n ${srv} nginx
  
}

teardown () {
  rio rm ${srv}
}

## Validation tests ##
@test "rio create - service creation, state & scale test" {
  rio ps
  [[ "$(rio ps --format '{{.Service.Name}}')" =~ ${srv} ]] || false
  [ "$(rio inspect --format '{{.name}}' ${srv})" == ${srv} ]
  rio --wait-state inactive wait ${srv}
  [ "$(rio inspect --format '{{.state}}' ${srv})" == "inactive" ]
  [ "$(rio inspect --format '{{.scale}}' ${srv})" == "0" ]

}

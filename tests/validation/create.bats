## Setup ##

setup() {
  export srv=tsrv${RANDOM}
  rio create -n ${srv} nginx
  
}

teardown () {
  rio rm ${srv}
}

## Validation tests ##
@test "rio create - service is listing" {
  rio ps
  [[ "$(rio ps --format '{{.Service.Name}}')" =~ ${srv} ]]
}

@test "rio create - service created" {
  rio ps
  [ "$(rio inspect --format '{{.name}}' ${srv})" == ${srv} ]
}

@test "rio create - service state inactive" {
  rio ps
  rio --wait-state inactive wait ${srv}
  [ "$(rio inspect --format '{{.state}}' ${srv})" == "inactive" ]
}

@test "rio create - service scale = 0" {
  rio ps
  [ "$(rio inspect --format '{{.scale}}' ${srv})" == "0" ]
}


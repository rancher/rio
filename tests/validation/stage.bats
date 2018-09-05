## Setup ##

setup() {
  export stk=tstk${RANDOM}
  export srv=tsrv${RANDOM}
  rio run -p 80/http --n ${stk}/${srv} --scale=3 ibuildthecloud/demo:v1
  rio stage --image=ibuildthecloud/demo:v3 ${stk}/${srv}:v3
  rio wait ${stk}/${srv}
  rio wait ${stk}/${srv}:v3
}

teardown () {
  rio rm ${stk}/${srv}
  rio stack rm ${stk}
}

## Validation tests ##
@test "rio stage - ensure v1 service is active" {
  #rio ps
  [[ "$(rio ps --format '{{.Service.Name}}')" =~ ${srv} ]] || false
}

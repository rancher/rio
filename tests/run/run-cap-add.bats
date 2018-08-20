## Setup ##

setup() {
  export stk=tstk${RANDOM}
  export srv=tsrv${RANDOM}
  rio stack create ${stk}
}

teardown () {
  rio rm ${stk}
}

capAddTestrio() {
  cmd="rio run -n ${stk}/${srv}"
  expect=""

  while [ $# -gt 0 ]; do
    cmd="${cmd} --cap-add $1"
    if [[ ! -z "${expect}" ]]; then
      expect="${expect} "
    fi
    expect="${expect}$1"
    shift
  done
  cmd="${cmd} nginx"

  $cmd
  rio wait ${stk}/${srv}

  nsp="$(rio inspect --format '{{.id}}' ${stk}/${srv} | cut -f1 -d:)"
  got="$(rio inspect --format '{{.capAdd}}' ${stk}/${srv})"
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [[ "${got}" == "[${expect}]" ]]
}

capAddTestk8s() {
  cmd="rio run -n ${stk}/${srv}"
  expect=""

  while [ $# -gt 0 ]; do
    cmd="${cmd} --cap-add $1"
    if [[ ! -z "${expect}" ]]; then
      expect="${expect},"
    fi
    expect="${expect}$1"
    shift
  done
  cmd="${cmd} nginx"

  $cmd
  rio wait ${stk}/${srv}

  nsp="$(rio inspect --format '{{.id}}' ${stk}/${srv} | cut -f1 -d:)"
  got=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r '.spec.template.spec.containers[0].securityContext.capabilities.add | join(",")')
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [[ "${got}" == "${expect}" ]]
}

@test "k8s ALL" {
  capAddTestk8s 'ALL'
}

@test "AUDIT CONTROL and SYSLOG" {
  capAddTestk8s 'AUDIT_CONTROL' 'SYSLOG'
}

@test "RIO ALL" {
  capAddTestrio 'ALL'
}

@test "RIO AUDIT CONTROL and SYSLOG" {
  capAddTestrio 'AUDIT_CONTROL' 'SYSLOG'
}


## Validation tests ##
@test "rio cap add - Not added" {
  #rio ps
  export srv=tsrv${RANDOM}
  rio run -n ${stk}/${srv} nginx
  rio wait ${stk}/${srv}
  [[ "$(rio inspect --format '{{.capAdd}}' ${stk}/${srv})" == "<no value>" ]]
}

@test "rio cap add - ALL" {
  #rio ps
  export srv=tsrv${RANDOM}
  rio run -n ${stk}/${srv} --cap-add ALL nginx
  rio wait ${stk}/${srv}
  [[ "$(rio inspect --format '{{.capAdd}}' ${stk}/${srv})" == "[ALL]" ]]
}

@test "rio cap add - multiple" {
  #rio ps
  export srv=tsrv${RANDOM}
  rio run -n ${stk}/${srv} --cap-add AUDIT_CONTROL --cap-add SYSLOG  nginx
  rio wait ${stk}/${srv}
  [[ "$(rio inspect --format '{{.capAdd}}' ${stk}/${srv})" == "[AUDIT_CONTROL SYSLOG]" ]]
}

@test "rio kubectl cap add - ALL" {
  #rio ps
  export srv=tsrv${RANDOM}
  rio run -n ${stk}/${srv} --cap-add ALL nginx
  rio wait ${stk}/${srv}
  nsp="$(rio inspect --format '{{.id}}' ${stk}/${srv} | cut -f1 -d:)"
  [[ "$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r .spec.template.spec.containers[0].securityContext.capabilities.add[0])" == "ALL"  ]]
}

@test "rio kubectl cap add - multiple" {
  #rio ps
  export srv=tsrv${RANDOM}
  rio run -n ${stk}/${srv} --cap-add ALL nginx
  rio wait ${stk}/${srv}
  nsp="$(rio inspect --format '{{.id}}' ${stk}/${srv} | cut -f1 -d:)"
  [[ "$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r .spec.template.spec.containers[0].securityContext.capabilities.add[0])" == "ALL"  ]]
}

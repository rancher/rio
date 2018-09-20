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
  cmd="${cmd} tfiduccia/counting"

  $cmd
  rio wait ${stk}/${srv}

  nsp="$(rio inspect --format '{{.id}}' ${stk}/${srv} | cut -f1 -d:)"
  got="$(rio inspect --format '{{.capAdd}}' ${stk}/${srv})"
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "[${expect}]" ]
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
  cmd="${cmd} tfiduccia/counting"

  $cmd
  rio wait ${stk}/${srv}

  nsp="$(rio inspect --format '{{.id}}' ${stk}/${srv} | cut -f1 -d:)"
  got=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r '.spec.template.spec.containers[0].securityContext.capabilities.add | join(",")')
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]
}

## Validation tests ##

@test "run capadd - ALL" {
  capAddTestrio 'ALL'
  capAddTestk8s 'ALL'

}

@test "rio run capadd - AUDIT CONTROL and SYSLOG" {
  capAddTestrio 'AUDIT_CONTROL' 'SYSLOG'
  capAddTestk8s 'AUDIT_CONTROL' 'SYSLOG'

}



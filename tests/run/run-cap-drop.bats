## Setup ##

setup() {
  export stk=tstk${RANDOM}
  export srv=tsrv${RANDOM}
  rio stack create ${stk}
}

teardown () {
  rio rm ${stk}
}

capDropTestrio() {
  cmd="rio run -n ${stk}/${srv}"
  expect=""

  while [ $# -gt 0 ]; do
    cmd="${cmd} --cap-drop $1"
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
  got="$(rio inspect --format '{{.capDrop}}' ${stk}/${srv})"
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "[${expect}]" ]
}

capDropTestk8s() {
  cmd="rio run -n ${stk}/${srv}"
  expect=""

  while [ $# -gt 0 ]; do
    cmd="${cmd} --cap-drop $1"
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
  got=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r '.spec.template.spec.containers[0].securityContext.capabilities.drop | join(",")')
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]
}

## Validation tests ##

@test "rio run capdrop - ALL" {
  capDropTestrio 'ALL'
  capDropTestk8s 'ALL'

}

@test "rio run capdrop - AUDIT CONTROL and SYSLOG" {
  capDropTestrio 'AUDIT_CONTROL' 'SYSLOG'
  capDropTestk8s 'AUDIT_CONTROL' 'SYSLOG'

}

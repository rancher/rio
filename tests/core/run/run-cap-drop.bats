## Setup ##

setup() {
  export stk=tstk${RANDOM}
  export srv=tsrv${RANDOM}
  rio stack create ${stk}
}

teardown () {
  rio rm ${stk}
}

runCapDroprio () {
  cmd="rio run -n ${stk}/${srv}"
  value=""

  while [ $# -gt 0 ]; do
    cmd="${cmd} --cap-drop $1"
    if [[ ! -z "${value}" ]]; then
      value="${value} "
    fi
    value="${value}$1"
    shift
  done
  cmd="${cmd} nginx"
  echo "cmd = ${cmd}"

  $cmd
  rio wait ${stk}/${srv}
}


capDropTestrio() {
  expect=""

  while [ $# -gt 0 ]; do
    if [[ ! -z "${expect}" ]]; then
      expect="${expect} "
    fi
    expect="${expect}$1"
    shift
  done

  got="$(rio inspect --format '{{.capDrop}}' ${stk}/${srv})"
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "[${expect}]" ]
}

capDropTestk8s() {
  expect=""
  i=0
  count=$#

  while [ $# -gt 0 ]; do
    expect="${expect},${1}"
    shift
  done


  nsp="$(rio inspect --format '{{.id}}' ${stk}/${srv} | cut -f1 -d:)"
  got=""
  
  while [ $i -lt $count ]; do
    filter=".spec.template.spec.containers[0].securityContext.capabilities.drop[${i}]"
    more=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r ${filter})
    got="${got},${more}"
  let i=$i+1
  done

   echo "Expect: ${expect}"
   echo "Got: ${got}"
  [ "${got}" == "${expect}" ]
}

## Validation tests ##

@test "run capdrop - SYSLOG" {
  runCapDroprio 'SYSLOG'
  capDropTestrio 'SYSLOG'
  capDropTestk8s 'SYSLOG'

}

@test "run capdrop - AUDIT CONTROL and SYSLOG" {
  runCapDroprio 'AUDIT_CONTROL' 'SYSLOG'
  capDropTestrio 'AUDIT_CONTROL' 'SYSLOG'
  capDropTestk8s 'AUDIT_CONTROL' 'SYSLOG'

}

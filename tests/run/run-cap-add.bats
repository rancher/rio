## Setup ##

setup() {
  export stk=tstk${RANDOM}
  export srv=tsrv${RANDOM}
  rio stack create ${stk}
}

teardown () {
  rio rm ${stk}
}

runCapAdd () {
  cmd="rio run -n ${stk}/${srv}"
  value=""

  while [ $# -gt 0 ]; do
    cmd="${cmd} --cap-add $1"
    if [[ ! -z "${value}" ]]; then
      value="${value} "
    fi
    value="${value}$1"
    shift
  done
  cmd="${cmd} tfiduccia/counting"

  echo $cmd
  $cmd
  rio wait ${stk}/${srv}
}

capAddTestrio() {
  expect=""

  while [ $# -gt 0 ]; do
    if [[ ! -z "${expect}" ]]; then
      expect="${expect} "
    fi
    expect="${expect}$1"
    shift
  done

  got="$(rio inspect --format '{{.capAdd}}' ${stk}/${srv})"
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "[${expect}]" ]
}

capAddTestk8s() {
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
    filter=".spec.template.spec.containers[0].securityContext.capabilities.add[${i}]"
    more=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r ${filter})
    got="${got},${more}"
  let i=$i+1
  done

   echo "Expect: ${expect}"
   echo "Got: ${got}"
  [ "${got}" == "${expect}" ]
}

## Validation tests ##

@test "run capadd - ALL" {
  runCapAdd 'ALL'
  capAddTestrio 'ALL'
  capAddTestk8s 'ALL'

}

@test "rio run capadd - AUDIT CONTROL and SYSLOG" {
  runCapAdd 'AUDIT_CONTROL' 'SYSLOG'
  capAddTestrio 'AUDIT_CONTROL' 'SYSLOG'
  capAddTestk8s 'AUDIT_CONTROL' 'SYSLOG'

}



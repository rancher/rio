#!/usr/bin/env bats
## Setup ##

setup() {
  export stk=tstk${RANDOM}
  export srv=tsrv${RANDOM}
  rio stack create ${stk}
}

teardown () {
  rio rm ${stk}
}

runEnvrio() {
  expect=""
  cmd="rio run -n ${stk}/${srv}"

  while [ $# -gt 0 ]; do
    cmd="${cmd} -e $1"
    if [[ ! -z "${expect}" ]]; then
      expect="${expect} "
    fi
    expect="${expect}$1"
    shift
  done
  cmd="${cmd} nginx"

  $cmd
}


capEnvTestrio() {
  expect=""

    while [ $# -gt 0 ]; do
    if [[ ! -z "${expect}" ]]; then
      expect="${expect} "
    fi
    expect="${expect}$1"
    shift
  done


  rio wait ${stk}/${srv}

  nsp="$(rio inspect --format '{{.id}}' ${stk}/${srv} | cut -f1 -d:)"
  got="$(rio inspect --format '{{.environment}}' ${stk}/${srv})"
  echo "Expect: [${expect}]"
  echo "Got: ${got}"
  [ "${got}" == "[${expect}]" ]
}

capEnvTestk8s() {
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
    filter=".spec.template.spec.containers[0].env[${i}] | join(\"=\")"
    more=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r "${filter}")
    got="${got},${more}"
    let i=$i+1
  done

  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]

}

## Validation tests ##

@test "run env - foo=bar" {
  runEnvrio 'foo=bar'
  capEnvTestrio 'foo=bar'
  capEnvTestk8s 'foo=bar'

}

@test "run env - foo=bar foo2=bar2" {
  runEnvrio 'foo=bar' 'foo2=bar2'
  capEnvTestrio 'foo=bar' 'foo2=bar2'
  capEnvTestk8s 'foo=bar' 'foo2=bar2'
  
}

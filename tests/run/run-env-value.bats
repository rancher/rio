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

capAddTestrio() {
  cmd="rio run -n ${stk}/${srv}"
  expect=""

  while [ $# -gt 0 ]; do
    cmd="${cmd} -e $1"
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
  got="$(rio inspect --format '{{.environment}}' ${stk}/${srv})"
  echo "Expect: [${expect}]"
  echo "Got: ${got}"
  [ "${got}" == "[${expect}]" ]
}

capAddTestk8s() {
  cmd="rio run -n ${stk}/${srv}"
  expect=""
  i=0
  count=$#

  while [ $# -gt 0 ]; do
    cmd="${cmd} -e $1"
    expect="${expect},${1}"
    shift
  done
  cmd="${cmd} tfiduccia/counting"

  $cmd
  rio wait ${stk}/${srv}


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

@test "rio foo=bar" {
  capAddTestrio 'foo=bar'
}

@test "rio foo=bar foo2=bar2" {
  capAddTestrio 'foo=bar' 'foo2=bar2'
}

@test "k8s foo=bar" {
  capAddTestk8s 'foo=bar'
}

@test "k8s foo=bar foo2=bar2" {
  capAddTestk8s 'foo=bar' 'foo2=bar2'
}





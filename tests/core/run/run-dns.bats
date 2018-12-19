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

runDnsrio() {
  value=""
  cmd="rio run -n ${stk}/${srv}"

  while [ $# -gt 0 ]; do
    cmd="${cmd} --dns $1"
    if [[ ! -z "${value}" ]]; then
      expect="${value} "
    fi
    expect="${value}$1"
    shift
  done
  cmd="${cmd} nginx"

  $cmd
  rio wait ${stk}/${srv}
}


dnsTestrio() {
  expect=""

    while [ $# -gt 0 ]; do
    if [[ ! -z "${expect}" ]]; then
      expect="${expect} "
    fi
    expect="${expect}$1"
    shift
  done


  rio wait ${stk}/${srv}

  got="$(rio inspect --format '{{.dns}}' ${stk}/${srv})"
  echo "Expect: [${expect}]"
  echo "Got: ${got}"
  [ "${got}" == "[${expect}]" ]
}

dnsTestk8s() {
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
    filter=".spec.template.spec.dnsConfig.nameservers[${i}]"
    more=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r "${filter}")
    got="${got},${more}"
    let i=$i+1
  done

  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]

}

## Validation tests ##

@test "run dns - 1.1.1.1" {
  runDnsrio '1.1.1.1'
  dnsTestrio '1.1.1.1'
  dnsTestk8s '1.1.1.1'

}

@test "run dns - 1.1.1.1 2.2.2.2" {
  runDnsrio '1.1.1.1' '2.2.2.2'
  dnsTestrio '1.1.1.1' '2.2.2.2'
  dnsTestk8s '1.1.1.1' '2.2.2.2'
  
}

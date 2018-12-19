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

runDnsOptionrio() {
  value=""
  cmd="rio run -n ${stk}/${srv}"

  while [ $# -gt 0 ]; do
    cmd="${cmd} --dns-option $1"
    if [[ ! -z "${value}" ]]; then
      expect="${value} "
    fi
    expect="${value}$1"
    shift
  done
  cmd="${cmd} nginx"

  $cmd
}


dnsOptionTestrio() {
  expect=""

    while [ $# -gt 0 ]; do
    if [[ ! -z "${expect}" ]]; then
      expect="${expect} "
    fi
    expect="${expect}$1"
    shift
  done


  rio wait ${stk}/${srv}

  got="$(rio inspect --format '{{.dnsOptions}}' ${stk}/${srv})"
  echo "Expect: [${expect}]"
  echo "Got: ${got}"
  [ "${got}" == "[${expect}]" ]
}

dnsOptionTestk8s() {
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
    filter=".spec.template.spec.dnsConfig.options[${i}] | join(\"=\")"
    more=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r "${filter}")
    got="${got},${more}"
    let i=$i+1
  done

  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]

}

## Validation tests ##

@test "run dns-option - debug" {
  runDnsOptionrio 'debug'
  dnsOptionTestrio 'debug'
  dnsOptionTestk8s 'debug'

}

@test "run dns-option - debug attempts:2" {
  runDnsOptionrio 'debug' 'attempts:2'
  dnsOptionTestrio 'debug' 'attempts:2'
  dnsOptionTestk8s 'debug' 'attempts:2'
  
}

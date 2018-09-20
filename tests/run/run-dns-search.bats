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

runDnsSearchrio() {
  value=""
  cmd="rio run -n ${stk}/${srv}"

  while [ $# -gt 0 ]; do
    cmd="${cmd} --dns-search $1"
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


dnsSearchTestrio() {
  expect=""

    while [ $# -gt 0 ]; do
    if [[ ! -z "${expect}" ]]; then
      expect="${expect} "
    fi
    expect="${expect}$1"
    shift
  done


  rio wait ${stk}/${srv}

  got="$(rio inspect --format '{{.dnsSearch}}' ${stk}/${srv})"
  echo "Expect: [${expect}]"
  echo "Got: ${got}"
  [ "${got}" == "[${expect}]" ]
}

dnsSearchTestk8s() {
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
    filter=".spec.template.spec.dnsConfig.searches[${i}]"
    more=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r "${filter}")
    got="${got},${more}"
    let i=$i+1
  done

  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]

}

## Validation tests ##

@test "run dns-search - example.com" {
  runDnsSearchrio 'example.com'
  dnsSearchTestrio 'example.com'
  dnsSearchTestk8s 'example.com'

}

@test "run dns-Search- example.com example2.com" {
  runDnsSearchrio 'example.com' 'example2.com'
  dnsSearchTestrio 'example.com' 'example2.com'
  dnsSearchTestk8s 'example.com' 'example2.com'
  
}

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

runEntpntrio() {
  cmd="rio run -n ${stk}/${srv}"
  value=""

  while [ $# -gt 0 ]; do
    cmd="${cmd} --entrypoint $1"
    if [[ ! -z "${value}" ]]; then
      expect="${value} "
    fi
    expect="${value}$1"
    shift
  done

  cmd="${cmd} nginx"
  $cmd
}


entpntTestrio() {
  expect=""

    while [ $# -gt 0 ]; do
    if [[ ! -z "${expect}" ]]; then
      expect="${expect} "
    fi
    expect="${expect}$1"
    shift
  done

  got="$(rio inspect --format '{{.entrypoint}}' ${stk}/${srv})"
  echo "Expect: [${expect}]"
  echo "Got: ${got}"
  [ "${got}" == "[${expect}]" ]
}


entpntTestk8s() {

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
    filter=".spec.template.spec.containers[0].command[${i}]"
    more=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r ${filter})
    got="${got},${more}"
  let i=$i+1
  done

   echo "Expect: ${expect}"
   echo "Got: ${got}"
  [ "${got}" == "${expect}" ]
  
}

@test "Entrypoint - set to sh" {
  entpnt="sh"
  runEntpntrio "${entpnt}"
  entpntTestrio "${entpnt}"
  entpntTestk8s "${entpnt}"

}

@test "Entrypoint - set to sh -i" {
  entpnt1="sh"
  entpnt2="-i"
  runEntpntrio "${entpnt1}" "${entpnt2}"
  entpntTestrio "${entpnt1}" "${entpnt2}"
  entpntTestk8s "${entpnt1}" "${entpnt2}"

}

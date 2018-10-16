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

runExposerio() {
  cmd="rio run -n ${stk}/${srv}"
  local value=""

  while [ $# -gt 0 ]; do
    cmd="${cmd} --expose $1"
    if [[ ! -z "${value}" ]]; then
      value="${value} "
    fi
    value="${value}$1"
    shift
  done
  cmd="${cmd} nginx"

  $cmd
  rio wait ${stk}/${srv}

}

exposeTestrio() {
  expect=""
  i=0
  count=$#
  local got=""

  while [ $# -gt 0 ]; do
    expect="${expect} ${1}"

    format="{{ (index .expose ${i}).targetPort }}"
    format2="{{ (index .expose ${i}).protocol }}"

    more="$(rio inspect --format "${format}" ${stk}/${srv})/$(rio inspect --format "${format2}" ${stk}/${srv})"
    got="${got} ${more}"
    
    let i=$i+1
    shift
  done

  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]

}



exposeTestk8s() {

  expect=""
  i=0
  count=$#
  local got=""
  local more=""

  while [ $# -gt 0 ]; do
    expect="${expect} ${1}"

    nsp="$(rio inspect --format '{{.id}}' ${stk}/${srv} | cut -f1 -d:)"
    
    filter1=".spec.template.spec.containers[0].ports[${i}].containerPort"
    filter2=".spec.template.spec.containers[0].ports[${i}].protocol"
    
    more=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r "${filter1}")"/"$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r "${filter2}")
    more=$(echo ${more} | awk '{print tolower($0)}')
    echo ${more}
    got="${got} ${more}"

    let i=$i+1
    shift
  done
  
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]

}


@test "run expose - port set to 22/tcp" {
  value="22/tcp"
  runExposerio "${value}"
  exposeTestrio "${value}"
  exposeTestk8s "${value}"

}

@test "run expose - port set to 22/tcp 80/udp" {
  value="22/tcp" 
  value2="80/udp"
  runExposerio "${value}" "${value2}"
  exposeTestrio "${value}" "${value2}"
  exposeTestk8s "${value}" "${value2}"
}

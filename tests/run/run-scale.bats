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

runScalerio() {
  cmd="rio run -n ${stk}/${srv}"
  value=$1

  cmd="${cmd} --scale ${value} nginx"
  ${cmd}
  
}

scaleTestrio() {
  expect=$1

  got="$(rio inspect --format '{{.scale}}' ${stk}/${srv})"
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]

}

scaleTestk8s() {
  expect=$1
  
  sleep 5

  nsp="$(rio inspect --format '{{.id}}' ${stk}/${srv} | cut -f1 -d:)"
  filter=".spec.replicas"
  got=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r "${filter}")
  
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]

}

@test "Scale - value set to 0" {
  scale=0
  runScalerio "${scale}"
  scaleTestrio "${scale}"
  scaleTestk8s "${scale}"

}

@test "Scale - value set to 1" {
  scale=1
  runScalerio "${scale}"
  scaleTestrio "${scale}"
  scaleTestk8s "${scale}"

}

@test "Scale - value set to 5" {
  scale=5
  runScalerio "${scale}"
  scaleTestrio "${scale}"
  scaleTestk8s "${scale}"

}

@test "Scale - value set to 10" {
  scale=10
  runScalerio "${scale}"
  scaleTestrio "${scale}"
  scaleTestk8s "${scale}"

}


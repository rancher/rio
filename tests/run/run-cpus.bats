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

runCpusrio() {
  cmd="rio run -n ${stk}/${srv}"
  value=$1

  cmd="${cmd} --cpus ${value} nginx"
  ${cmd}
  
}

cpusTestrio() {
  expect=$1

  got="$(rio inspect --format '{{.nanoCpus}}' ${stk}/${srv})"
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]

}

cpusTestk8s() {
  expect=$1
  

  nsp="$(rio inspect --format '{{.id}}' ${stk}/${srv} | cut -f1 -d:)"
  filter=".spec.template.spec.containers[0].resources.requests.cpu"
  got=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r "${filter}")
  
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]

}

@test "cpus - value set to 0" {
  cpus=0
  runCpusrio "${cpus}"
  cpusTestrio "${cpus}"
  cpusTestk8s "${cpus}"

}

@test "cpus - value set to 1" {
  cpus=1
  runCpusrio "${cpus}"
  cpusTestrio "${cpus}"
  cpusTestk8s "${cpus}"

}

@test "cpus - value set to 5" {
  cpus=5
  runCpusrio "${cpus}"
  cpusTestrio "${cpus}"
  cpusTestk8s "${cpus}"

}

@test "cpus - value set to 10" {
  cpus=10
  runCpusrio "${cpus}"
  cpusTestrio "${cpus}"
  cpusTestk8s "${cpus}"

}


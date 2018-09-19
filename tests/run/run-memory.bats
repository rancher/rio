## Setup ##

setup() {
  export stk=tstk${RANDOM}
  export srv=tsrv${RANDOM}
  rio stack create ${stk}
}

teardown () {
  rio rm ${stk}
}

runMemoryrio() {
  cmd="rio run -n ${stk}/${srv}"
  size=$1
  unit=$2

  cmd="${cmd} --memory ${size}${unit} nginx"
  ${cmd}
  
}

memoryTestrio() {
  expect=$1

  got=$(rio inspect --format "{{.memoryReservationBytes | json}}" ${stk}/${srv})
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]

}

memoryTestk8s() {
  expect=$1

  rio wait ${stk}/${srv}

  nsp="$(rio inspect --format '{{.id}}' ${stk}/${srv} | cut -f1 -d:)"
  filter=".spec.template.spec.containers[0].resources.requests.memory"
  got=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r "${filter}")
  
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]

}

@test "memory reservation - test byte default 100000000" {
  runMemoryrio "100000000" ""
  memoryTestrio "100000000"
  memoryTestk8s "100M"
}

@test "memory reservation - test 100000000b" {
  runMemoryrio "100000000" "b"
  memoryTestrio "100000000"
  memoryTestk8s "100M"
}

@test "memory reservation - test 100000k" {
  runMemoryrio "100000" "k"
  memoryTestrio "102400000"
  memoryTestk8s "102400k"
}

@test "memory reservation - test 10m" {
  runMemoryrio "10" "m"
  memoryTestrio "10485760"
  memoryTestk8s "10485760"
}

@test "memory reservation - test 1g" {
  runMemoryrio "1" "g"
  memoryTestrio "1073741824"
  memoryTestk8s "1073741824"
}




    


















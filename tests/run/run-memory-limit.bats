## Setup ##

setup() {
  export stk=tstk${RANDOM}
  export srv=tsrv${RANDOM}
  rio stack create ${stk}
}

teardown () {
  rio rm ${stk}
}

runMemoryLimitrio() {
  cmd="rio run -n ${stk}/${srv}"
  size=$1
  unit=$2

  cmd="${cmd} --memory-limit ${size}${unit} nginx"
  ${cmd}
  
}

memoryLimitTestrio() {
  expect=$1

  got=$(rio inspect --format "{{.memoryLimitBytes | json}}" ${stk}/${srv})
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]

}

memoryLimitTestk8s() {
  expect=$1

  rio wait ${stk}/${srv}

  nsp="$(rio inspect --format '{{.id}}' ${stk}/${srv} | cut -f1 -d:)"
  filter=".spec.template.spec.containers[0].resources.limits.memory"
  got=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r "${filter}")
  
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]

}

@test "memory limit - test byte default 100000000" {
  runMemoryLimitrio "100000000" ""
  memoryLimitTestrio "100000000"
  memoryLimitTestk8s "100M"
}

@test "memory limit - test 100000000b" {
  runMemoryLimitrio "100000000" "b"
  memoryLimitTestrio "100000000"
  memoryLimitTestk8s "100M"
}

@test "memory limit - test 100000k" {
  runMemoryLimitrio "100000" "k"
  memoryLimitTestrio "102400000"
  memoryLimitTestk8s "102400k"
}

@test "memory limit - test 10m" {
  runMemoryLimitrio "10" "m"
  memoryLimitTestrio "10485760"
  memoryLimitTestk8s "10485760"
}

@test "memory limit - test 1g" {
  runMemoryLimitrio "1" "g"
  memoryLimitTestrio "1073741824"
  memoryLimitTestk8s "1073741824"
}




    


















## Setup ##

setup() {
  export stk=tstk${RANDOM}
  export srv=tsrv${RANDOM}
  rio stack create ${stk}
}

teardown () {
  rio rm ${stk}
}

createAddConfigFile() {

  cfile=$(mktemp -t rio-test.XXXXX)

  while [ $# -gt 0 ]; do
    echo $1 >> ${cfile}
    shift
  done   

  export config=tconfig${RANDOM}
  rio config create ${stk}/${config} ${cfile}
  rm ${cfile}
}

runConfigrio() {
  cmd="rio run -n ${stk}/${srv}"
  cmd="${cmd} --config ${config}:/temp nginx"
  $cmd
  rio wait ${stk}/${srv}
}


configTestrio() {
  expect=$1
  field=$2

  format="{{ (index .configs 0).${field} }}"
  got=$(rio inspect --format "${format}" ${stk}/${srv})
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]
}

configTestk8s() {
  expect=$1
  field=$2
  i=0

  nsp="$(rio inspect --format '{{.id}}' ${stk}/${srv} | cut -f1 -d:)"
  filter=".spec.template.spec.containers[0].volumeMounts[0].${field}"
  got=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r "${filter}")
  
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]

}



## Validation tests ##


@test "rio config - validate target" {
  createAddConfigFile "foo=bar" "foo2=bar2"
  runConfigrio ""
  configTestrio "/temp" "target"
  configTestk8s "/temp" "mountPath"

}

@test "rio config - validate source" {
  createAddConfigFile "foo=bar" "foo2=bar2"
  runConfigrio ""
  configTestrio "${config}" "source"
  configTestk8s "config-${config}" "name"

}


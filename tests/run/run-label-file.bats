## Setup ##

setup() {
  export stk=tstk${RANDOM}
  export srv=tsrv${RANDOM}
  rio stack create ${stk}
}

teardown () {
  rio rm ${stk}
  rm ${lfile}
}

createAddLabelFile() {

  export lfile=$(mktemp -t rio-test.XXXXX)

  while [ $# -gt 0 ]; do
    echo $1 >> ${lfile}
    shift
  done

}

runLabelrio(){
  cmd="rio run -n ${stk}/${srv}"
  cmd="${cmd} --label-file ${lfile} nginx"
  $cmd
  
  rio wait ${stk}/${srv}

}

labelFileTestrio() {
  key=$1
  expect=$2

  nsp="$(rio inspect --format '{{.id}}' ${stk}/${srv} | cut -f1 -d:)"
  filter=".spec.template.metadata.labels.${key}"
  got=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r "${filter}")

  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]
}

labelFileTestk8s() {
  key=$1
  expect=$2

  nsp="$(rio inspect --format '{{.id}}' ${stk}/${srv} | cut -f1 -d:)"
  filter=".spec.template.metadata.labels.${key}"
  got=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r "${filter}")
  
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]

}


## Validation tests ##


@test "rio file of labels - validate single label" {
  createAddLabelFile "foo1=bar1"
  runLabelrio ""
  labelFileTestrio "foo1" "bar1"
  labelFileTestk8s "foo1" "bar1"

}

@test "rio file of labels - validate multiple labels" {
  createAddLabelFile "foo1=bar1" "foo2=bar2"
  runLabelrio ""
  labelFileTestrio "foo1" "bar1"
  labelFileTestrio "foo2" "bar2"
  labelFileTestk8s "foo1" "bar1"
  labelFileTestk8s "foo2" "bar2"
  
}


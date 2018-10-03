## Setup ##

setup() {
  export stk=tstk${RANDOM}
  export srv=tsrv${RANDOM}
  rio stack create ${stk}
}

teardown () {
  rio rm ${stk}
}

runMetadatario(){
  cmd="rio run -n ${stk}/${srv}"
  value=""

  while [ $# -gt 0 ]; do
    cmd="${cmd} --metadata $1"
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

metadataTestrio() {
  key=$1
  expect=$2
  
  filter=".metadata.${key}"
  got=$(rio inspect --format "{{$filter}}" ${stk}/${srv})
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]
}

metadataTestk8s() {
  key=$1
  expect=$2

  nsp="$(rio inspect --format '{{.id}}' ${stk}/${srv} | cut -f1 -d:)"
  filter=".spec.template.metadata.annotations.${key}"
  got=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r "${filter}")
  
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]

}


## Validation tests ##


@test "run metadata - validate single metadata value" {
  runMetadatario "foo1=bar1"
  metadataTestrio "foo1" "bar1"
  metadataTestk8s "foo1" "bar1"

}

@test "run metadata - validate multiple metadata values" {
  runMetadatario "foo1=bar1" "foo2=bar2"
  metadataTestrio "foo1" "bar1"
  metadataTestrio "foo2" "bar2"
  metadataTestk8s "foo1" "bar1"
  metadataTestk8s "foo2" "bar2"

}

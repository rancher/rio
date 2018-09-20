## Setup ##

setup() {
  export stk=tstk${RANDOM}
  export srv=tsrv${RANDOM}
  rio stack create ${stk}
}

teardown () {
  rio rm ${stk}
}


runLabelrio(){
  cmd="rio run -n ${stk}/${srv}"
  value=""

  while [ $# -gt 0 ]; do
    cmd="${cmd} --label $1"
    if [[ ! -z "${value}" ]]; then
      value="${value} "
    fi
    value="${value}$1"
    shift
  done
  cmd="${cmd} tfiduccia/counting"

  $cmd
  rio wait ${stk}/${srv}

}

labelTestrio() {
  key=$1
  expect=$2
  
  filter=".labels.${key}"
  got=$(rio inspect --format "{{$filter}}" ${stk}/${srv})
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]
}

labelTestk8s() {
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


@test "rio label - validate single label" {
  runLabelrio "foo1=bar1"
  labelTestrio "foo1" "bar1"
  labelTestk8s "foo1" "bar1"
}

@test "rio label - validate multiple labels" {
  runLabelrio "foo1=bar1" "foo2=bar2"
  labelTestrio "foo1" "bar1"
  labelTestrio "foo2" "bar2"
  labelTestk8s "foo1" "bar1"
  labelTestk8s "foo2" "bar2"

}

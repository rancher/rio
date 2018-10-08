## Setup ##

setup() {
  export stk=tstk${RANDOM}
  export srv=tsrv${RANDOM}
  rio stack create ${stk}
}

teardown () {
  rio rm ${stk}
}

runImagerio() {
  cmd="rio run -n ${stk}/${srv}"
  value=$1

  cmd="${cmd} --image-pull-policy ${value} nginx"
  $cmd
  rio wait ${stk}/${srv}

}
 


imagePullTestrio() {
  value=$1

  got=$(rio inspect --format '{{.imagePullPolicy}}' ${stk}/${srv})
  echo "Expect: ${value}"
  echo "Got: ${got}"
  [ "${got}" == "${value}" ]
}

imagePullTestk8s() {
  value=$1
  expect=$2

  nsp="$(rio inspect --format '{{.id}}' ${stk}/${srv} | cut -f1 -d:)"
  filter=".spec.template.spec.containers[0].imagePullPolicy"
  got=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r "${filter}" )
  
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]

}


## Validation tests ##


@test "rio image-pull-policy - default" {
    rio run -n ${stk}/${srv} nginx
    got=$(rio inspect --format '{{.imagePullPolicy}}' ${stk}/${srv})
    [ "${got}" = "not-present" ]
}

@test "run image-pull-policy - always" {
  runImagerio "always"
  imagePullTestrio "always"
  imagePullTestk8s "always" "Always"

}

@test "run image-pull-policy - never" {
  runImagerio "never"
  imagePullTestrio "never"
  imagePullTestk8s "never" "Never"

}

@test "run image-pull-policy - not-present" {
  runImagerio "not-present"
  imagePullTestrio "not-present"
  imagePullTestk8s "not-present" "IfNotPresent"

}

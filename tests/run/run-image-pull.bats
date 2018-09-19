## Setup ##

setup() {
  export stk=tstk${RANDOM}
  export srv=tsrv${RANDOM}
  rio stack create ${stk}
}

teardown () {
  rio rm ${stk}
}


imagePullTestrio() {
  cmd="rio run -n ${stk}/${srv}"
  value=$1

  cmd="${cmd} --image-pull-policy ${value} nginx"
  $cmd
  rio wait ${stk}/${srv}

  got=$(rio inspect --format '{{.imagePullPolicy}}' ${stk}/${srv})
  echo "Expect: ${value}"
  echo "Got: ${got}"
  [ "${got}" == "${value}" ]
}

imagePullTestk8s() {
  cmd="rio run -n ${stk}/${srv}"
  value=$1
  expect=$2

  cmd="${cmd} --image-pull-policy ${value} nginx"
  $cmd
  rio wait ${stk}/${srv}

  nsp="$(rio inspect --format '{{.id}}' ${stk}/${srv} | cut -f1 -d:)"
  filter=".spec.template.spec.containers[0].imagePullPolicy"
  got=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r "${filter}" )
  
  echo "Expect: ${expect}"
  echo "Got: ${got}"
  [ "${got}" == "${expect}" ]

}


## Validation tests ##


@test "rio image pull policy - default" {
    rio run -n ${stk}/${srv} nginx
    got=$(rio inspect --format '{{.imagePullPolicy}}' ${stk}/${srv})
    [ "${got}" = "not-present" ]
}

@test "rio image pull policy - always" {
  imagePullTestrio "always"
}

@test "rio image pull policy - never" {
  imagePullTestrio "never"
}

@test "rio image pull policy - not-present" {
  imagePullTestrio "not-present"
}

@test "k8s image pull policy - always" {
  imagePullTestk8s "always" "Always"
}

@test "k8s image pull policy - never" {
  imagePullTestk8s "never" "Never"
}

@test "k8s image pull policy - not-present" {
  imagePullTestk8s "not-present" "IfNotPresent"
}

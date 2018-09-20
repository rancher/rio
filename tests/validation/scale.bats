## Setup ##

setup() {
  export srv=tsrv${RANDOM}
  rio run -n ${srv} nginx
  rio scale ${srv}=3
  rio wait ${srv}
}

teardown () {
  rio rm ${srv}
}

## Validation tests ##


@test "scale - rio scale & k8s replica check" {
  rio ps
  [ "$(rio inspect --format '{{.scale}}' ${srv})" == "3" ]
  nsp="$(rio inspect --format '{{.id}}' ${srv} | cut -f1 -d:)"
  [ "$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r .status.replicas)" == "3" ]

}
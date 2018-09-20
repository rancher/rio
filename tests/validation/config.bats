## Setup ##

setup() {
  export config=tconfig${RANDOM}
  echo "foo=bar" > config.txt
  rio config create ${config} config.txt
}

teardown () {
  rio config rm ${config}
  rm config.txt
}

## Validation tests ##
@test "rio config - config is listing" {
  rio config
  [ "$(rio inspect --format '{{.name}}' ${config})" == ${config} ]
}

@test "config - contents are correct" {
  rio config
  [ "$(rio inspect --format '{{.content}}' ${config})" == "foo=bar" ]
  nsp="$(rio inspect --format '{{.id}}' ${config} | cut -f1 -d:)"
  [ "$(rio kubectl get config -n ${nsp} -o=json ${config} | jq -r .spec.content)" == "foo=bar" ]

}

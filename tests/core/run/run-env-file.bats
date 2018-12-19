## Setup ##

setup() {
  export stk=tstk${RANDOM}
  export srv=tsrv${RANDOM}
  rio stack create ${stk}
}

teardown () {
  rio rm ${stk}
  rm ${efile}

}

createEnvFile() {

  export efile=$(mktemp -t rio-test.XXXXX)

  while [ $# -gt 0 ]; do
    echo $1 >> ${efile}
    shift
  done

}

runEnvFilerio() {
  cmd="rio run -n ${stk}/${srv}"
  cmd="${cmd} --env-file ${efile} nginx"
  echo "command: ${cmd}"
  $cmd

  rio wait ${stk}/${srv}
}


envFileTestrio() {
  expect=""

    while [ $# -gt 0 ]; do
    if [[ ! -z "${expect}" ]]; then
      expect="${expect} "
    fi
    expect="${expect}$1"
    shift
  done


  rio wait ${stk}/${srv}

  got="$(rio inspect --format '{{.environment}}' ${stk}/${srv})"
  echo "Expect: [${expect}]"
  echo "Got: ${got}"
  [ "${got}" == "[${expect}]" ]

}

envFileTestk8s() {
  expect=""
  i=0
  count=$#

  while [ $# -gt 0 ]; do
    expect="${expect},${1}"
    shift
  done


  nsp="$(rio inspect --format '{{.id}}' ${stk}/${srv} | cut -f1 -d:)"
  got=""
  
  while [ $i -lt $count ]; do
    filter=".spec.template.spec.containers[0].env[${i}].name"
    filter2=".spec.template.spec.containers[0].env[${i}].value"
    more=$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r ${filter})
    more=${more}"="$(rio kubectl get -n ${nsp} -o=json deploy/${srv} | jq -r ${filter2})
    got="${got},${more}"
  let i=$i+1
  done

   echo "Expect: ${expect}"
   echo "Got: ${got}"
  [ "${got}" == "${expect}" ]
}


## Validation tests ##


@test "run environment file - foo=bar" {
  value1="foo=bar"
  createEnvFile ${value1}
  runEnvFilerio ""
  envFileTestrio ${value1}
  envFileTestk8s ${value1}

}

@test "run environment file - foo=bar, foo2=bar2" {
  value1="foo=bar"
  value2="foo2=bar2"
  createEnvFile ${value1} ${value2}
  runEnvFilerio ""
  envFileTestrio ${value1} ${value2}
  envFileTestk8s ${value1} ${value2}

}


## Setup ##

setup() {
  export stack=tstack${RANDOM}
  export srv=tsrv${RANDOM}
  export domain=d${RANDOM}
}

teardown () {
  rio domain rm ${domain}
  rio rm ${stack}
}

@test "public domain - service target" {
    rio run -p 80/http -n ${stack}/${srv} nginx:latest
    rio wait ${stack}/${srv}
    rio domain add --space default --stack ${stack} --service ${srv} --domain test.foo.bar ${domain}
    ns=$(rio inspect ${stack}/${srv} | jq .id | tr -d '"' | cut -d':' -f1)
    result=$(rio kubectl get virtualservice ${srv} -n ${ns} -o json | jq '.spec.hosts[2]' | tr -d '"')
    [[ ${result} == "test.foo.bar" ]]
}
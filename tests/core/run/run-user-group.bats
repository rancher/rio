#!/usr/bin/env bats
## setup ##

setup() {
  export stk=tstk${RANDOM}
}

teardown () {
  rio rm ${stk}
}

@test "run user group" {
  srv=tsrv${RANDOM}
  rio run -n ${stk}/${srv} --user 10:10 alpine
  [[ $(rio inspect ${stk}/${srv} | jq .group | tr -d "\"") == "10" ]] || false
  [[ $(rio inspect ${stk}/${srv} | jq .user | tr -d "\"") == "10" ]] || false

  srv=tsrv${RANDOM}
  rio run -n ${stk}/${srv} --user 10 --group 10 alpine
  [[ $(rio inspect ${stk}/${srv} | jq .group | tr -d "\"") == "10" ]] || false
  [[ $(rio inspect ${stk}/${srv} | jq .user | tr -d "\"") == "10" ]] || false

  srv=tsrv${RANDOM}
  rio run -n ${stk}/${srv} --user 10 alpine
  [[ $(rio inspect ${stk}/${srv} | jq .user | tr -d "\"") == "10" ]] || false

  srv=tsrv${RANDOM}
  rio run -n ${stk}/${srv} --group 10 alpine
  [[ $(rio inspect ${stk}/${srv} | jq .group | tr -d "\"") == "10" ]] || false
}

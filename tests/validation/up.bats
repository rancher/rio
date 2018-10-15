setup() {
    export stack_name=foo-${RANDOM}
}

teardown() {
    rio rm $stack_name
}

## Validation tests ##
@test "rio up- remote repository" {
    rio up $stack_name https://raw.githubusercontent.com/StrongMonkey/rio-compose-files/master/test1/test1-stack.yaml
    rio wait $stack_name/my-test
    length=$(rio inspect ${stack_name}/my-test | jq '.configs | length')
    [ $length == "2" ]
}
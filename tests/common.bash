
setup() {
    STACK="test-st-${RANDOM}"
    rio stack create $STACK
}

teardown() {
    rio rm --type stack $STACK
}

#!/usr/bin/env bats

load common

@test "set weight to zero" {
    n=$STACK/test
    v2=${n}:v2
    rio run --name $n busybox
    rio stage $v2
    rio weight ${v2}=20%
    [ "$(rio inspect --format "{{.revisions.v2.weight}}" $v2)" == "20" ]
    rio weight ${v2}=0
    [ "$(rio inspect --format "{{.revisions.v2.weight}}" $v2)" == "0" ]
}

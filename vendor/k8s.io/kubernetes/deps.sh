#!/bin/bash
set -e

git rm -rf test
find -name '*_test.go' -exec git rm {} \;
find -depth -name testdata -type d -exec git rm -rf {} \;
find -depth -name testing -type d -exec git rm -rf {} \;

cat << EOF | sed -E 's!^([^/]+/[^/]+/[^/]+)(/[^ ]+) (.*)!\1 \3!g' | sed -E 's!^((google|[ckv])[^/]+/[^/]+)(/[^ ]+) (.*)!\1 \4!g' > vendor.conf
package=k8s.io/kubernetes
package=k8s.io/kubernetes/cmd/hyperkube
$(cat ./Godeps/Godeps.json | jq -r '(.Deps | .[] | "\(.ImportPath) \(.Comment) \(.Rev)\n")' | sed 's/null//' | awk '{print $1 " " $2}' | grep -Ev 'github.com/opencontainers/runc|bitbucket.org/ww/goautoneg|github.com/google/cadvisor' | sort -k2,1 | uniq -f1)
bitbucket.org/ww/goautoneg       a547fc61f48d567d5b4ec6f8aee5573d8efce11d  https://github.com/rancher/goautoneg.git
github.com/ibuildthecloud/kvsql  c649f12fe5250718e4e024b8b40e7de796ab095e
github.com/google/cadvisor       87e237ff35b9d752ba58860a06e0ebe57816cbb7  https://github.com/ibuildthecloud/cadvisor.git
github.com/opencontainers/runc   96ec2177ae841256168fcf76954f7177af9446eb
EOF

trash
git rm -rf Godeps
rm trash.lock
cd vendor/k8s.io
ln -s ../../staging/src/k8s.io/* .
cd ../..
git add vendor vendor.conf
for i in ./cmd/*/*.go; do
    echo Building $(dirname $i)
    go build $(dirname $i)
done
git commit -m "Update vendor"

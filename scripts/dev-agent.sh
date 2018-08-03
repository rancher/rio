#!/bin/bash
set -e

cd $(dirname $0)/../bin

echo Compiling Agent
go build -tags k3s -o ../image/agent ../agent/main.go

echo Compiling CLI
go build -tags k3s -o rio-agent ../cli/main.go

echo Building image
../image/build

echo Running
exec sudo ENTER_ROOT=../image/main.squashfs ./rio-agent --debug agent -s https://localhost:7443 -t $(<${HOME}/.rancher/rio/server/node-token)

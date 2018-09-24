#!/bin/bash
set -e

cd $(dirname $0)/../bin

echo Compiling
go build -tags k3s -o rio-k8s-server ../cli/main.go

echo Running
exec ./rio-k8s-server --debug server --disable-agent

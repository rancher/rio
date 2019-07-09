#!/usr/bin/env bash
refdocs -config ./apidocs/doc-config.json -api-dir "github.com/rancher/rio/pkg/apis/" -out-file ./api-docs.md --template-dir ./apidocs

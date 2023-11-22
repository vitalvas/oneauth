#!/bin/bash

set -e

mkdir -p build

GOOS=darwin GOARCH=amd64 go build -ldflags '-w -s' -o build/oneauth_darwin_amd64 cmd/oneauth/main.go

tar -czvf build/oneauth_darwin_amd64.tar.gz -C build/ oneauth_darwin_amd64

if [ ! -z "${GITHUB_ACTIONS}" ]; then
    aws s3 cp build/oneauth_darwin_amd64.tar.gz s3://vv-github-build-artifacts/${GITHUB_REPOSITORY}/${GITHUB_REF_NAME}/${GITHUB_SHA}/oneauth_darwin_amd64.tar.gz
fi

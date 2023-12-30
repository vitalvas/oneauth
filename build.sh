#!/bin/bash

set -e

GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)

mkdir -p build/${GOOS}/${GOARCH}

VERSION="v0.0.$(date +%s)"
LD_FLAGS="-w -s -X \"github.com/vitalvas/oneauth/internal/buildinfo.Version=${VERSION}\""

## -- Build the agent binary --
go build -ldflags "${LD_FLAGS}" -o build/${GOOS}/${GOARCH}/oneauth cmd/oneauth/main.go

# ensure the binary is working
./build/${GOOS}/${GOARCH}/oneauth --version

tar -czvf build/oneauth_${GOOS}_${GOARCH}.tar.gz -C build/${GOOS}/${GOARCH} oneauth

if [ ! -z "${GITHUB_ACTIONS}" ]; then
    aws s3 cp build/oneauth_${GOOS}_${GOARCH}.tar.gz s3://vv-github-build-artifacts/${GITHUB_REPOSITORY}/${GITHUB_REF_NAME}/${GITHUB_SHA}/oneauth_${GOOS}_${GOARCH}.tar.gz
fi

## -- Build the server binary --
if [ "${GOOS}" == "linux" ] || [ "${GOOS}" == "darwin" ]; then
    CGO_ENABLED=0 go build -ldflags "${LD_FLAGS}" -o build/${GOOS}/${GOARCH}/oneauth-server cmd/server/main.go

    # ensure the binary is working
    ./build/${GOOS}/${GOARCH}/oneauth-server --version

    tar -czvf build/oneauth-server_${GOOS}_${GOARCH}.tar.gz -C build/${GOOS}/${GOARCH} oneauth-server

    if [ ! -z "${GITHUB_ACTIONS}" ]; then
        aws s3 cp build/oneauth-server_${GOOS}_${GOARCH}.tar.gz s3://vv-github-build-artifacts/${GITHUB_REPOSITORY}/${GITHUB_REF_NAME}/${GITHUB_SHA}/oneauth-server_${GOOS}_${GOARCH}.tar.gz
    fi
fi

## -- Build the ssh-test-server --
if [ "${GOOS}" == "linux" ]; then
    CGO_ENABLED=0 go build -ldflags "${LD_FLAGS}" -o build/${GOOS}/${GOARCH}/oneauth-ssh-test-server cmd/ssh-test-server/main.go

    # ensure the binary is working
    ./build/${GOOS}/${GOARCH}/oneauth-ssh-test-server --version

    tar -czvf build/oneauth-ssh-test-server_${GOOS}_${GOARCH}.tar.gz -C build/${GOOS}/${GOARCH} oneauth-ssh-test-server

    if [ ! -z "${GITHUB_ACTIONS}" ]; then
        aws s3 cp build/oneauth-ssh-test-server_${GOOS}_${GOARCH}.tar.gz s3://vv-github-build-artifacts/${GITHUB_REPOSITORY}/${GITHUB_REF_NAME}/${GITHUB_SHA}/oneauth-ssh-test-server_${GOOS}_${GOARCH}.tar.gz
    fi
fi

#!/usr/bin/env bash

MY_PATH="`dirname \"$0\"`" # relative bash file path
DAPP_PROXY_DIR="`( cd \"$MY_PATH/..\" && pwd )`"  # absolutized and normalized dappctrl path

GIT_COMMIT=$(git rev-list -1 HEAD)
GIT_RELEASE=$(git tag -l --points-at HEAD)

cd "${DAPP_PROXY_DIR}"

echo
echo go get
echo

go get -u -v github.com/rakyll/statik

echo
echo go generate
echo

go generate -x ./...

echo
echo go build
echo

echo $GOPATH/bin/dappproxy
go build -o $GOPATH/bin/dappproxy -ldflags "-X main.Commit=$GIT_COMMIT \
    -X main.Version=$GIT_RELEASE" -tags=notest \
    ${DAPP_PROXY_DIR}/adapter || exit 1

echo
echo done

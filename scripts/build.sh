#!/usr/bin/env bash

MY_PATH="`dirname \"$0\"`" # relative bash file path
DAPP_PROXY_DIR="`( cd \"$MY_PATH/..\" && pwd )`"  # absolutized and normalized dapp-proxy path

GIT_COMMIT=$(git rev-list -1 HEAD)
GIT_RELEASE=$(git tag -l --points-at HEAD)

# if $GIT_RELEASE is zero:
GIT_RELEASE=${GIT_RELEASE:-$(git rev-parse --abbrev-ref HEAD | grep -o "[0-9]\{1,\}\.[0-9]\{1,\}\.[0-9]\{1,\}")}

echo
echo GIT_COMMIT=${GIT_COMMIT}
echo GIT_RELEASE=${GIT_RELEASE}

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

go build -o $GOPATH/bin/dappproxy -ldflags "-X main.Commit=$GIT_COMMIT \
    -X main.Version=$GIT_RELEASE" -tags=notest \
    ${DAPP_PROXY_DIR}/adapter || exit 1
echo $GOPATH/bin/dappproxy

echo
echo done

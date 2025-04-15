#!/usr/bin/env ksh
set -v

build() {
    os=$1
    arch=$2
    GOOS=$1 GOARCH=$2 go build -ldflags="-s -w" -o bin/$os-$arch/savefile
}

cd $(dirname $0)

if [[ ! -e bin ]]; then
    mkdir bin
fi

go fmt ./...
go vet ./...
build linux amd64
build darwin amd64
build darwin arm64
build windows amd64

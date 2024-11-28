#!/bin/bash -xe

destination=$1
version=$(grep "^go " go.mod |awk '{print $2}')

arch="$(uname -m)"

case $arch in
    x86_64 | amd64)
        arch="amd64"
        ;;
    aarch64 | arm64)
        arch="arm64"
        ;;
    s390x)
        arch="s390x"
        ;;
    *)
        echo "ERROR: invalid arch=${arch}, only support x86_64, aarch64 and s390x"
        exit 1
        ;;
esac

tarball=go$version.linux-$arch.tar.gz
url=https://dl.google.com/go/

mkdir -p $destination
curl -L $url/$tarball -o $destination/$tarball
tar -xvf $destination/$tarball -C $destination

#! /bin/sh
# cross-compile magnacarto for various os/cpus and zip/tar.gz the output
# requires Go 1.5, set GO15 environment var to use a different Go installation

set -e

if [ ! -z $GO15 ]; then
    export PATH=${GO15}/bin:$PATH
    export GOROOT=${GO15}
fi

BUILD_DATE=`date +%Y%m%d`
BUILD_REF=`git rev-parse --short HEAD`
BUILD_VERSION=dev-$BUILD_DATE-$BUILD_REF

VERSION_LDFLAGS="-X github.com/omniscale/magnacarto.buildVersion=${BUILD_VERSION}"

# build os arch
function build() {
    os=$1
    arch=$2
    build_name=magnacarto-$BUILD_VERSION-$os-$arch
    mkdir -p $build_name
    echo building $build_name
    cd $build_name
    env GOOS=$os GOARCH=$arch godep go build -ldflags "$VERSION_LDFLAGS" github.com/omniscale/magnacarto/cmd/magnacarto
    env GOOS=$os GOARCH=$arch godep go build -ldflags "$VERSION_LDFLAGS" github.com/omniscale/magnacarto/cmd/magnaserv
    cp ../../../{README.md,LICENSE} ./
    cd ..
    if [ $os = windows ]; then
        zip -r $build_name.zip $build_name
    else
        tar -cvzf $build_name.tar.gz $build_name
    fi
    rm -r $build_name
}

mkdir -p dist/$BUILD_VERSION
cd dist/$BUILD_VERSION

# build for these os/arch combinations
build windows 386
build windows amd64
build linux 386
build linux amd64
build darwin amd64

cd ../../

#! /bin/sh
# cross-compile magnacarto for various os/cpus and zip/tar.gz the output
# requires Go 1.5, set GO15 environment var to use a different Go installation

set -e

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
    env GOOS=$os GOARCH=$arch go build -ldflags "$VERSION_LDFLAGS" github.com/omniscale/magnacarto/cmd/magnacarto
    env GOOS=$os GOARCH=$arch go build -ldflags "$VERSION_LDFLAGS" github.com/omniscale/magnacarto/cmd/magnaserv
    # use git archive to only include checked-in files
    (cd ../../../ && git archive --format tar HEAD app README.md LICENSE) | tar -x -
    (cd ../../../docs && git archive --format tar HEAD examples/) | tar -x -
    cd ..
    if [ $os = windows ]; then
        rm -f $build_name.zip
        zip -q -r $build_name.zip $build_name
    else
        tar -czf $build_name.tar.gz $build_name
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

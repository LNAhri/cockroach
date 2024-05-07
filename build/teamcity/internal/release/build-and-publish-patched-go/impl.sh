#!/usr/bin/env bash

set -xeuo pipefail

# When updating to a new Go version, update all of these variables.
GOVERS=1.19.13
GOLINK=https://go.dev/dl/go$GOVERS.src.tar.gz
SRCSHASUM=ccf36b53fb0024a017353c3ddb22c1f00bc7a8073c6aac79042da24ee34434d3
# We mirror the upstream freebsd because we don't have a cross-compiler targeting it.
GOFREEBSDLINK=https://go.dev/dl/go$GOVERS.freebsd-amd64.tar.gz
FREEBSDSHASUM=97fd4990c5349ab922b9bf3e4c655e899135559ea6ad666d4b3c7a27b1e147a2

apt-get update
DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    apt-transport-https \
    ca-certificates \
    clang-10 \
    cmake \
    curl \
    git \
    gnupg2 \
    golang \
    make \
    python-is-python3 \
    python3 \
    python3.8-venv

update-alternatives --install /usr/bin/clang clang /usr/bin/clang-10 100 \
    --slave /usr/bin/clang++ clang++ /usr/bin/clang++-10

curl -fsSL $GOFREEBSDLINK -o /artifacts/go$GOVERS.freebsd-amd64.tar.gz
echo "$FREEBSDSHASUM  /artifacts/go$GOVERS.freebsd-amd64.tar.gz" | sha256sum -c -

# libtapi is required for later versions of MacOSX.
git clone https://github.com/tpoechtrager/apple-libtapi.git
cd apple-libtapi
git checkout a66284251b46d591ee4a0cb4cf561b92a0c138d8
./build.sh
./install.sh
cd ..
rm -rf apple-libtapi

curl -fsSL https://storage.googleapis.com/public-bazel-artifacts/toolchains/crosstool-ng/x86_64/20220711-205918/aarch64-unknown-linux-gnu.tar.gz -o aarch64-unknown-linux-gnu.tar.gz
echo '58407f1f3ed490bd0a0a500b23b88503fbcc25f0f69a0b7f8a3e8e7b9237341b aarch64-unknown-linux-gnu.tar.gz' | sha256sum -c -
curl -fsSL https://storage.googleapis.com/public-bazel-artifacts/toolchains/osxcross/x86_64/20220317-165434/x86_64-apple-darwin21.2.tar.gz -o x86_64-apple-darwin21.2.tar.gz
echo '751365dbfb5db66fe8e9f47fcf82cbbd7d1c176b79112ab91945d1be1d160dd5 x86_64-apple-darwin21.2.tar.gz' | sha256sum -c -
curl -fsSL https://storage.googleapis.com/public-bazel-artifacts/toolchains/crosstool-ng/x86_64/20220711-205918/x86_64-unknown-linux-gnu.tar.gz -o x86_64-unknown-linux-gnu.tar.gz
echo '8b0c246c3ebd02aceeb48bb3d70c779a1503db3e99be332ac256d4f3f1c22d47 x86_64-unknown-linux-gnu.tar.gz' | sha256sum -c -
curl -fsSL https://storage.googleapis.com/public-bazel-artifacts/toolchains/crosstool-ng/x86_64/20220711-205918/x86_64-w64-mingw32.tar.gz -o x86_64-w64-mingw32.tar.gz
echo 'b87814aaeed8c68679852029de70cee28f96c352ed31c4c520e7bee55999b1c6 x86_64-w64-mingw32.tar.gz' | sha256sum -c -
echo *.tar.gz | xargs -n1 tar -xzf
rm *.tar.gz

curl -fsSL $GOLINK -o golang.tar.gz
echo "$SRCSHASUM  golang.tar.gz" | sha256sum -c -
mkdir -p /tmp/go$GOVERS
tar -C /tmp/go$GOVERS -xzf golang.tar.gz
rm golang.tar.gz
cd /tmp/go$GOVERS/go
# NB: we apply a patch to the Go runtime to keep track of running time on a
# per-goroutine basis. See #82356 and #82625.
git apply /bootstrap/diff.patch /bootstrap/6446af9.diff /bootstrap/008-fix-CVE-2023-45288.patch /bootstrap/a79ea27.diff
cd ..

CONFIGS="linux_amd64 linux_arm64 darwin_amd64 darwin_arm64 windows_amd64"

for CONFIG in $CONFIGS; do
    case $CONFIG in
        linux_amd64)
            CC_FOR_TARGET=/x-tools/x86_64-unknown-linux-gnu/bin/x86_64-unknown-linux-gnu-cc
            CXX_FOR_TARGET=/x-tools/x86_64-unknown-linux-gnu/bin/x86_64-unknown-linux-gnu-c++
        ;;
        linux_arm64)
            CC_FOR_TARGET=/x-tools/aarch64-unknown-linux-gnu/bin/aarch64-unknown-linux-gnu-cc
            CXX_FOR_TARGET=/x-tools/aarch64-unknown-linux-gnu/bin/aarch64-unknown-linux-gnu-c++
        ;;
        darwin_amd64)
            CC_FOR_TARGET=/x-tools/x86_64-apple-darwin21.2/bin/x86_64-apple-darwin21.2-cc
            CXX_FOR_TARGET=/x-tools/x86_64-apple-darwin21.2/bin/x86_64-apple-darwin21.2-c++
            ;;
        darwin_arm64)
            CC_FOR_TARGET=/x-tools/x86_64-apple-darwin21.2/bin/aarch64-apple-darwin21.2-cc
            CXX_FOR_TARGET=/x-tools/x86_64-apple-darwin21.2/bin/aarch64-apple-darwin21.2-c++
        ;;
        windows_amd64)
            CC_FOR_TARGET=/x-tools/x86_64-w64-mingw32/bin/x86_64-w64-mingw32-cc
            CXX_FOR_TARGET=/x-tools/x86_64-w64-mingw32/bin/x86_64-w64-mingw32-c++
        ;;
    esac
    GOOS=$(echo $CONFIG | cut -d_ -f1)
    GOARCH=$(echo $CONFIG | cut -d_ -f2)
    cd go/src
    if [ $GOOS == darwin ]; then
        export LD_LIBRARY_PATH=/x-tools/x86_64-apple-darwin21.2/lib
    fi
    GOOS=$GOOS GOARCH=$GOARCH CC=clang CXX=clang++ CC_FOR_TARGET=$CC_FOR_TARGET CXX_FOR_TARGET=$CXX_FOR_TARGET \
               GOROOT_BOOTSTRAP=$(go env GOROOT) CGO_ENABLED=1 ./make.bash
    if [ $GOOS == darwin ]; then
        unset LD_LIBRARY_PATH
    fi
    cd ../..
    rm -rf /tmp/go$GOVERS/go/pkg/${GOOS}_$GOARCH/cmd
    for OTHER_CONFIG in $CONFIGS; do
        if [ $CONFIG != $OTHER_CONFIG ]; then
            rm -rf /tmp/go$GOVERS/go/pkg/$OTHER_CONFIG
        fi
    done
    if [ $CONFIG != linux_amd64 ]; then
        rm go/bin/go go/bin/gofmt
        mv go/bin/${GOOS}_$GOARCH/* go/bin
        rm -r go/bin/${GOOS}_$GOARCH
    fi
    tar cf - go | gzip -9 > /artifacts/go$GOVERS.$GOOS-$GOARCH.tar.gz
    rm -rf go/bin
done

sha256sum /artifacts/*.tar.gz

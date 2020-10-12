#!/bin/bash

# Explicitly assign the /builds/luajit as install
# directory for luajit.
export PREFIX=/builds/luajit
sudo mkdir -p $PREFIX
sudo chmod a+rwx $PREFIX

# Initialize the environments when they are absent.
if [ "$MAKE" == "" ]; then export MAKE=make; fi
if [ "$CC" == "" ]; then export CC=gcc; fi
if [ "$CXX" == "" ]; then export CXX=g++; fi
if [ "$GOOS" == "" ]; then export GOOS=linux; fi
if [ "$GOARCH" == "" ]; then export GOARCH=amd64; fi
export CC="$CC -L$PWD"
export CXX="$CC -L$PWD"

# Extraly add $PWD/ndk/bin to the path if there's
# such directory present.
if [ -d "$PWD/ndk/bin" ]; then
  export PATH="$PATH:$PWD/ndk/bin";

  # XXX(aegisudio): there's a setting in go that
  # enforces a specification of -pthread, which
  # causes a compile error. To negate the error,
  # we create a pseudo archive for it.
  ar m $PWD/libpthread.a
fi

# Print out variables for debugging purpose.
echo "Current Shell \$PATH: $PATH"
echo "Current Go Environment: "
CGO_ENABLED=1 GO111MODULE=on go env

# Build the LuaJIT (checked-out in the workflow).
echo "Build LuaJIT for link dependencies"
cd luajit
if [ "$HOST_CC" == "" ]; then
$MAKE BUILDMODE=static CC="$CC" \
      PREFIX=/builds/luajit clean all install
else
$MAKE BUILDMODE=static HOST_CC="$HOST_CC" \
      XCFLAGS+="$XCFLAGS" CROSS="$CROSS" \
      TARGET_SYS="$TARGET_SYS" FILE_T="$FILE_T" \
      PREFIX=/builds/luajit clean all install
fi
cd -

# Build the client connector archive.
echo "Build the TechminoOnline client connector"
CGO_ENABLED=1 GO111MODULE=on GOPROXY=https://goproxy.io \
  go build -x -ldflags "-w -s -extldflags \"-L$PWD\"" \
  -buildmode="c-shared" -o "$1" -v \
  -tags 'osusergo netgo' ./cmd/client

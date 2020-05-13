#!/bin/bash

# Initialize the environments when they are absent.
if [ "$MAKE" == "" ]; then export MAKE=make; fi
if [ "$CC" == "" ]; then export CC=gcc; fi
if [ "$CXX" == "" ]; then export CXX=g++; fi
if [ "$GOOS" == "" ]; then export GOOS=linux; fi
if [ "$GOARCH" == "" ]; then export GOARCH=amd64; fi

# Build the LuaJIT (checked-out in the workflow).
echo "Build LuaJIT for link dependencies"
cd luajit
$MAKE BUILDMODE=static CC="$CC" clean all install
cd -

# Build the client connector archive.
echo "Build the TechminoOnline client connector"
CGO_ENABLED=1 GO111MODULE=on GOPROXY=https://goproxy.io \
	go build -ldflags '-w -s' -buildmode="c-shared" \
	-o "$1" -v ./cmd/client

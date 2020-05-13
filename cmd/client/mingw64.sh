#!/bin/sh

set -e
PATH="/mingw64/bin:$PATH"

# Install MinGW build packages dependencies.
echo "Install MinGW depdencies using MinGW pacman"
pacman -S --noconfirm mingw-w64-x86_64-gcc \
                      mingw-w64-x86_64-make \
                      mingw-w64-x86_64-go \
                      mingw-w64-x86_64-pkg-config

# Build the LuaJIT (checked-out in the workflow).
echo "Build LuaJIT for link dependencies"
cd luajit
mingw32-make BUILDMODE=static clean all install
cd -

# Build the client connector archive.
echo "Build the TechminoOnline client connector"
PKG_CONFIG_PATH="/usr/local/lib/pkgconfig" \
GO111MODULE=on GOPROXY=https://goproxy.io go build \
	-ldflags '-w -s' -buildmode="c-shared" \
	-o client.dll -v ./cmd/client

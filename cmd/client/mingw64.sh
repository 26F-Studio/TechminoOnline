#!/bin/sh

set -e
export PATH="/mingw64/bin:$PATH"
export GOROOT="/mingw64/lib/go"

# Install MinGW build packages dependencies.
echo "Install MinGW depdencies using MinGW pacman"
pacman -S --noconfirm mingw-w64-x86_64-gcc \
                      mingw-w64-x86_64-make \
                      mingw-w64-x86_64-go \
                      mingw-w64-x86_64-pkg-config

# Setup environment and execute the build shell.
export MAKE=mingw32-make
export GOOS=windows
export PKG_CONFIG_PATH="/usr/local/lib/pkgconfig"
$(dirname $0)/build.sh client.dll

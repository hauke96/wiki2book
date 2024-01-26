#!/bin/bash

# Get latest tag for this commit
VERSION=$(git describe --tags --abbrev=0)

function build()
{
	OS=$1
	ARCH=$2
	SUFFIX=$3
	BIN_NAME="wiki2book-$VERSION-$ARCH-$OS$SUFFIX"

	echo "Build for $OS on $ARCH"
	# The -ldflags "-s -w" parameter makes the binary smaller by not generating symbol table and debugging information.
	GOOS=$OS GOARCH=$ARCH go build -ldflags "-s -w" -o ../$BIN_NAME .
}

(
	cd src

	if [[ "$1" == "all" ]]
	then
		build windows amd64 ".exe"
		build linux amd64
		build darwin amd64
	else
		build linux amd64
	fi
)
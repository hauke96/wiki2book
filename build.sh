#!/bin/bash

# Get latest tag for this commit
VERSION=$(grep --color=never "VERSION = " src/util/constants.go | grep --color=never -Po "v[\d\.]+")
GIVEN_OUTPUT=""

# Some default value
OS=all
ARCH=amd64

function usage()
{
	cat <<EOF
Usage: $0 -a <arch> -o <os> [-f <output-file>] [-h]

Parameter:
  -a  Architecture of the system as golang uses them (e.g. amd64, arm64). Default: amd64.
  -o  Operating system as golang uses them (e.g. windows, linux, darwin). Or use "all" to build for all operating systems. Default: all.
  -f  Optional: Output file. If not given, then a filename including the version, arch and os is chosen. Will be ignored when operating system is "all".
  -h  Prints this message.
EOF
}

function build()
{
	OS=$1
	ARCH=$2
	OUTPUT=$3

	if [[ $OS == "windows" ]]
	then
		OUTPUT="$OUTPUT.exe"
	fi

	echo "Build for $OS with $ARCH arch to $GIVEN_OUTPUT"

	# The -ldflags "-s -w" parameter makes the binary smaller by not generating symbol table and debugging information.
	GOOS=$OS GOARCH=$ARCH go build -ldflags "-s -w" -o $OUTPUT .
}

while getopts "a:o:f:h" opt; do
	case "$opt" in
		a)
			ARCH=${OPTARG}
			;;
		o)
			OS=${OPTARG}
			;;
		f)
			GIVEN_OUTPUT="$OPTARG"
			;;
		h)
			usage
			exit 0
			;;
		*)
			echo
			usage
			exit 0
			;;
	esac
done
shift $((OPTIND-1))

if [[ $ARCH =~ ^.*(x86_64|x64|aarch64).*$ ]]
then
	echo "Unknown architecture '$ARCH' given, but it looks like an 64-bit arch. I'll use 'amd64' (golang parameter for x86_64 / x64 architecture)."
	ARCH="amd64"
fi

if [[ "$GIVEN_OUTPUT" != "" ]]
then
	OUTPUT=$(realpath "$GIVEN_OUTPUT")
else
	OUTPUT=$(realpath "wiki2book-$VERSION-$OS-$ARCH")
fi

(
	cd src
	if [[ $OS == "all" ]]
	then
		build "windows" $ARCH $(realpath "wiki2book-$VERSION-windows-$ARCH")
		build "linux" $ARCH $(realpath "wiki2book-$VERSION-linux-$ARCH")
		build "darwin" $ARCH $(realpath "wiki2book-$VERSION-darwin-$ARCH")
	else
		build $OS $ARCH $OUTPUT
	fi
)

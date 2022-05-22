#!/bin/bash

if [ "$1" == "" ] || [ "$1" == "--help" ]
then
	echo "Please provide a test name."
	echo
	echo "Example usage to update \"test-foo\":"
	echo "    ./update.sh foo"
	exit 1
fi

echo "Run run.sh to generate files"
./run.sh

NAME="$1"

echo "Copy files"
cp "results/test-$NAME/test-$NAME.filelist" .
cp "results/test-$NAME/test-$NAME.html" .

echo "Done"
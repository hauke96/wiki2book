#!/bin/bash

if [ "$1" == "" ]
then
	echo "Please provide a test name."
	echo
	echo "Example usage to update \"test-foo\":"
	echo "    ./update.sh foo"
	exit 1
fi

echo "Run run.sh to generate files"
./run.sh > /dev/null 2>&1

NAME="$1"

echo "Copy files"
cp "results/test-$NAME/test-$NAME.filelist" .
cp "results/test-$NAME/test-$NAME.html" .

echo "Done"
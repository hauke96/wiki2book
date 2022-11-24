#!/bin/bash

if [ "$1" == "" ] || [ "$1" == "--help" ]
then
	echo "Please provide at least one test name."
	echo
	echo "Example usage:"
	echo "    ./update.sh foo bar blubb"
	echo
	echo "Example usage with running the test before updating:"
	echo "    ./update.sh -r foo bar blubb"
	exit 1
fi

if [ "$1" == "-r" ]
then
	echo "Run run.sh to generate files"
	./run.sh
fi

for NAME in "$@"
do
	echo "Copy files for test-$NAME"
	cp "results/test-$NAME/test-$NAME.filelist" .
	cp "results/test-$NAME/test-$NAME.html" .
done

echo "Done"

#!/bin/bash

LOGS="./logs"

# Build project
echo "Build project..."

cd ../src
go build .
mv src ../test/wiki2book

# Go back into test directory
cd ../test

echo "Building project done"
echo

# Create empty directories
echo "Prepare directories"
rm -rf $LOGS
mkdir $LOGS

echo "Preparing directories done"
echo

echo "Start tests:"
echo

function run()
{
	# $1 - Test title (e.g. "foo" for "test-foo.mediawiki" test file)

	START=`date +%s`
	echo "$1: Start"

	# TODO create own style and cover files for these integration tests
	./wiki2book standalone -o "test-$1" -s ../example/style.css -c ../example/wikipedia-astronomie-cover.png "test-$1.mediawiki" > "$LOGS/$1.log" 2>&1

	diff -q "test-$1/test-$1.html" "test-$1.html" > /dev/null
	if [ $? -ne 0 ]
	then
		echo "$1: FAIL"
		echo "$1: HTML differs:"
		git diff --no-index "test-$1/test-$1.html" "test-$1.html"
	else
		echo "$1: Success"
	fi

	END=`date +%s`
	echo "$1: Finished after `expr $END - $START` seconds"
	echo
}

# Run tests
run generic

echo "Finished all tests"

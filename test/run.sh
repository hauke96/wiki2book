#!/bin/bash

# Build project
echo "Build project..."

cd ../src
go build .
mv src ../test/wiki2book

# Go back into test directory
cd ../test

echo "Building project done."
echo

# Create empty directories
echo "Prepare directories."
rm -rf logs
mkdir logs

echo "Preparing directories done."
echo

echo "Start tests:"
echo

function run()
{
	# $1 - Test title (e.g. "foo" for "test-foo.mediawiki" test file)

	START=`date +%s`
	echo "Run test $1"

	./wiki2book --file $1

	END=`date +%s`
	echo "Finished test $1 after `expr $END - $START` seconds"
}

# Run tests
run generic

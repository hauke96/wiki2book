#!/bin/bash

LOGS="./logs"
FAILED_TESTS=""

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

	OUT="results/test-$1"

	START=`date +%s`
	echo "$1: Start"

	# TODO create own style and cover files for these integration tests
	./wiki2book standalone -o "$OUT" -s ../example/style.css -c ../example/wikipedia-astronomie-cover.png "test-$1.mediawiki" > "$LOGS/$1.log" 2>&1

	diff -q "$OUT/test-$1.html" "test-$1.html" > /dev/null
	if [ $? -ne 0 ]
	then
		echo "$1: FAIL"
		echo "$1: HTML differs:"
		git --no-pager diff --no-index "$OUT/test-$1.html" "test-$1.html"
		FAILED_TESTS+=" $1"
	else
		echo "$1: Success"
	fi

	END=`date +%s`
	echo "$1: Finished after `expr $END - $START` seconds"
	echo
}

# Run tests
PREFIX="test-"
SUFFIX=".mediawiki"
for f in $(find *.mediawiki)
do
	F=${f%"$SUFFIX"}
	run ${F#"$PREFIX"}
done

echo "Finished all tests"
echo

if [ "$FAILED_TESTS" != "" ]
then
	echo "These tests FAILED:"
	for t in "$FAILED_TESTS"
	do
		echo "    $t"
	done
fi

#!/bin/bash

LOGS="./logs"   # Folder with log files for each test
FAILED_TESTS="" # List of test names that failed

# Build project
echo "Build project..."

cd ../src
go build .
mv src ../test/wiki2book

echo "Building project done"
echo

# Go back into test directory
cd ../test

# Create empty log-directory
echo "Prepare log directory"
rm -rf $LOGS
mkdir $LOGS

echo "Preparing log directory done"
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

	# Generate and check file list (except the .epub file which will always have a different hash value)
	find $OUT -type f -exec sha256sum {} \; | grep -v "\.epub" > "$OUT/test-$1.filelist"
	diff -q "$OUT/test-$1.filelist" "test-$1.filelist" > /dev/null
	if [ $? -ne 0 ]
	then
		echo "$1: FAIL"
		echo "$1: Files differ:"
		git --no-pager diff --no-index "$OUT/test-$1.filelist" "test-$1.filelist"
		FAILED_TESTS+=" $1"
		echo "$1: Some of the file differences might have been caused by Wikipedia (e.g. when the math rendering changes slightly)"
	fi

	# Compare HTML files
	diff -q "$OUT/test-$1.html" "test-$1.html" > /dev/null
	if [ $? -ne 0 ]
	then
		echo "$1: FAIL"
		echo "$1: HTML differs:"
		git --no-pager diff --no-index "$OUT/test-$1.html" "test-$1.html"
		FAILED_TESTS+=" $1"
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

# If test failed, list them
if [ "$FAILED_TESTS" != "" ]
then
	echo "These tests FAILED:"
	for t in "$FAILED_TESTS"
	do
		echo "    $t"
	done
else
	echo "All tests ran SUCCESSFULLY :)"
fi

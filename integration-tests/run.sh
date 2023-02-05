#!/bin/bash

GLOBAL_START=$(($(date +%s%N)/1000000))

HOME=$PWD
LOGS="./logs"   # Folder with log files for each test
FAILED_TESTS_WITH_CAUSE="" # List of test names with the fail-cause (e.g. "[HTML]")
FAILED_TESTS="" # List of failed test names only

# Build project
echo "Build project..."

cd ../src
go build .
mv src "$HOME/wiki2book"

echo "Building project done"
echo

# Go back into test directory
cd $HOME

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
	TEST_FAILED=0

	OUT="results/test-$1"
	mkdir -p "$OUT"

	START=$(($(date +%s%N)/1000000))
	echo "$1: Start test"

	# TODO create own style and cover files for these integration tests
	./wiki2book -d standalone -r --cache-dir "$OUT" -o "$OUT" -s ./style.css -c ./cover.png "test-$1.mediawiki" > "$LOGS/$1.log" 2>&1

	EXIT_CODE=$?
	if [ $EXIT_CODE -ne 0 ]
	then
		echo "$1: FAIL"
		echo "$1: wiki2book exited with code $EXIT_CODE"
		FAILED_TESTS_WITH_CAUSE+="$1 [exit-code-$EXIT_CODE]"$'\n'
		TEST_FAILED=1
	else
		# Generate and check file list
		find $OUT -type f | LC_ALL=C sort > "$OUT/test-$1.filelist"
		diff -q "test-$1.filelist" "$OUT/test-$1.filelist" > /dev/null
		if [ $? -ne 0 ]
		then
			echo "$1: FAIL"
			echo "$1: Files differ:"
			git --no-pager diff --no-index "test-$1.filelist" "$OUT/test-$1.filelist"
			FAILED_TESTS_WITH_CAUSE+="$1 [filelist]"$'\n'
			TEST_FAILED=1
			echo "$1: Some of the file differences might have been caused by Wikipedia (e.g. when the math rendering changes slightly)"
		fi

		# Compare HTML files
		diff -q "test-$1.html" "$OUT/test-$1.html" > /dev/null
		if [ $? -ne 0 ]
		then
			echo "$1: FAIL"
			echo "$1: HTML differs:"
			git --no-pager diff --no-index "test-$1.html" "$OUT/test-$1.html"
			FAILED_TESTS_WITH_CAUSE+="$1 [HTML]"$'\n'
			TEST_FAILED=1
		fi
	fi

	if [ $TEST_FAILED -ne 0 ]
	then
#		FAILED_TESTS_WITH_CAUSE+="\n"
		FAILED_TESTS+="$1 "
	fi

	END=$(($(date +%s%N)/1000000))
	echo "$1: Finished after `expr $END - $START` milliseconds"
}

# Run tests
PREFIX="test-"
SUFFIX=".mediawiki"
echo "=========="
for f in $(find *.mediawiki)
do
	F=${f%"$SUFFIX"}
	run ${F#"$PREFIX"}
	echo "=========="
done

GLOBAL_END=$(($(date +%s%N)/1000000))

echo
echo "Finished all integration-tests after `expr $GLOBAL_END - $GLOBAL_START` milliseconds"
echo

# If test failed, list them
if [ "$FAILED_TESTS_WITH_CAUSE" != "" ]
then
	echo "These integration-tests FAILED:"
	IFS=$'\n'
	for t in $FAILED_TESTS_WITH_CAUSE
	do
		echo -n "    "
#		echo $t | sed "s/\n/\n    /g"
		echo $t
	done

	echo "To update all failed test, run the update script with the following parameters:"
	echo "    $FAILED_TESTS"
else
	echo "Integration-tests ran SUCCESSFULLY :)"
fi

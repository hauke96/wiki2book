#!/bin/bash

# Use this instead, if you also want to know the test coverage:
#   go test -coverprofile test.out ./...
#   go tool cover -html=test.out

OUT=$(go test -parallel 1 -cover ./...)

if [[ $? != 0 ]]
then
	echo "$OUT"
	echo
	echo "Unit tests FAILED!"
	exit 1
fi

TEST_RESULTS=""
PKGS_WITHOUT_TESTS=""
while read -r line
do
	NAME=$(echo $line | grep -o "wiki2book[a-zA-Z/_]*" | sed 's/wiki2book\///g')
	if [[ $NAME == "" ]]
	then
		NAME="-"
	fi

	if [[ $line != ok* ]]
	then
		PKGS_WITHOUT_TESTS="$PKGS_WITHOUT_TESTS\n$NAME"
		continue
	fi

	TIME=$(echo $line | grep -o "[0-9]*\.[0-9]*s")
	if [[ $TIME == "" ]]
	then
		TIME="-"
	fi

	COVERAGE=$(echo $line | grep -o "coverage: [0-9.]*%")
	if [[ $COVERAGE == "" ]]
	then
		COVERAGE="-"
	fi

	TEST_RESULTS="$TEST_RESULTS\n$NAME $TIME $COVERAGE"
done <<< $OUT

echo "Test and coverage results:"
column -t <<< $(echo -e $TEST_RESULTS) | sed 's/^/    /g'

if [[ $PKGS_WITHOUT_TESTS != "" ]]
then
	echo
	echo -n "Packages without tests:"
	echo -e $PKGS_WITHOUT_TESTS | sed 's/^/    /g'
fi

echo
echo "Units tests were SUCCESSFUL!"

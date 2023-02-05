#!/bin/bash

# Use this instead, if you also want to know the test coverage:
#   go test -coverprofile test.out ./...
#   go tool cover -html=test.out

go test ./...

if [[ $? != 0 ]]
then
	echo
	echo "Unit tests FAILED!"
	exit 1
fi

echo
echo "Units tests were SUCCESSFUL!"

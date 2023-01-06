#!/bin/bash

cd src
go test ./...

if [[ $? != 0 ]]
then
	echo
	echo "Unit tests FAILED!"
	exit 1
fi

cd ../integration-tests
./run.sh

#!/bin/bash

function hline {
	echo
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' =
	echo " $1"
    printf '%*s\n' "${COLUMNS:-$(tput cols)}" '' | tr ' ' =
	echo
}

echo "Unit and integration tests will run now. This might take some time depending on the caches and internet speed."

hline "Unit tests"

cd src
./test.sh
UNIT_TEST_EXIT_CODE=$?
cd ..

hline "Integration tests"

cd ./integration-tests
./run.sh
INTEG_TEST_EXIT_CODE=$?

hline "Summary"

if [[ $UNIT_TEST_EXIT_CODE == 0 ]]
then
	echo "Unit tests:        OK"
else
	echo "Unit tests:        FAIL"
fi

if [[ $INTEG_TEST_EXIT_CODE == 0 ]]
then
	echo "Integration tests: OK"
else
	echo "Integration tests: FAIL"
fi

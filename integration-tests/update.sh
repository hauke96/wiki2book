#!/bin/bash

if [[ "$1" == "" ]] || [[ $@ == *--help* ]]
then
	cat <<EOF
Provide at least one test name to update.

Example to update the two given tests:

  ./update.sh real-article-Erde generic

Example usage with running the test before updating:

  ./update.sh -r real-article-Erde generic

EOF
	exit 1
fi

if [[ $@ == *-r* ]]
then
	echo "Run run.sh to generate files"
	./run.sh
fi

for NAME in "$@"
do
	if [[ $NAME == -* ]]
	then
		# ignore CLI parameters
		continue
	fi

	echo "Copy files for test-$NAME"
	cp "results/test-$NAME/test-$NAME.filelist" .
	cp "results/test-$NAME/html/test-$NAME.html" .
done

echo "Done"

#!/bin/bash

#PKGS_TO_IGNORE="github.com/alecthomas/kong,github.com/hauke96/sigolo,github.com/go-shiori/go-epub,github.com/pkg/errors,golang.org/x/net,os,filepath,strings,runtime"

(
	echo "Go into src"
	cd src

	# go install github.com/kisielk/godepgraph@latest
	#echo "Generate dot file with godepgraph"
	#OUT=$(godepgraph -s -p $PKGS_TO_IGNORE . | sed 's/wiki2book\///g')
	#echo "$OUT" > graph.dot
	#echo "Turn dot file into PNG"
	#echo "$OUT" | dot -Tpng -o graph.png

	# go install github.com/ofabry/go-callvis@latest
	#echo "Generate dot file with go-callvis"
	#OUT=$(go-callvis -algo static -format dot -file graph2 -group type -ignore $PKGS_TO_IGNORE -minlen 5 .)
	#echo "Turn dot file into PNG"
	#cat graph2.dot | dot -Tpng -o graph2.png

	# go install golang.org/x/tools/cmd/callgraph@latest
	echo "Generate dot file with callgraph"
	# Filter by call within wiki2book and exclude some very basic functions/packages that are simply called too often
	OUT=$(
		callgraph -algo static -format graphviz . | \
		grep -P "^( ? ?\".?.?wiki2book.* -> \".?.?wiki2book|digraph.*|})" | \
		grep -v -P "wiki2book/(util)" | \
		grep -vi -P "(\.init|print|config)" | \
		grep -v -P "parser\.Tokenizer\)\.(getToken|setRawToken)\"" | \
		sed 's/wiki2book\///g' | \
		uniq
	)

	echo "Generate full graph"
	echo "$OUT" > graph-full.dot
	ls -alh graph-full.dot
	echo "Turn dot file into PNG"
	cat graph-full.dot | dot -Tpng -o graph-full.png

	echo "Generate high-level graph"
	echo "$OUT" | grep -v -P "\".?.?(generator|parser|http|wikipedia)\." > graph-high-level.dot
	ls -alh graph-high-level.dot
	echo "Turn dot file into PNG"
	cat graph-high-level.dot | dot -Tpng -o graph-high-level.png

	echo "Generate graph with parser package"
	echo "$OUT" | grep -v -P "\".?.?(generator|http|wikipedia|wiki2book|image)[/.]" | grep -v "findCorrespondingCloseToken" > graph-parser.dot
	ls -alh graph-parser.dot
	echo "Turn dot file into PNG"
	cat graph-parser.dot | dot -Tpng -o graph-parser.png

	echo "Generate graph with generator package"
	echo "$OUT" | grep -v -P "\".?.?(parser|http|wikipedia|wiki2book)[/.]" > graph-generator.dot
	ls -alh graph-generator.dot
	echo "Turn dot file into PNG"
	cat graph-generator.dot | dot -Tpng -o graph-generator.png
)

echo "Done"
#| sed 's/github.com\/hauke96\/wiki2book\/server\///g' \
#| sed 's/github.com\/hauke96\/wiki2book\/server/main/g' \
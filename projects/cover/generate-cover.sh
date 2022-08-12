#!/bin/bash

set -e
set -x

rm -rf .tmp
mkdir -p .tmp

cp cover.tex logo.png .tmp
cd .tmp

ESCAPED_TITLE=${1//\\/\\\\}
ESCAPED_PRE=${2//\\/\\\\}
ESCAPED_PST=${3//\\/\\\\}
sed -i "s/((TITLE))/$ESCAPED_TITLE/" cover.tex
sed -i "s/((PRE))/$ESCAPED_PRE/" cover.tex
sed -i "s/((POST))/$ESCAPED_POST/" cover.tex

pdflatex ./cover.tex
convert -background white -alpha remove cover.pdf cover.png

mv cover.png ../
cd ../

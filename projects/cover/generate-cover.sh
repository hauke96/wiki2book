#!/bin/bash

set -e

POSITIONAL_ARGS=()

# Logo and template files will be copied, so we need defaults here
LOGO="./logo.png"
TEMPLATE="./cover.tex"

while [[ $# -gt 0 ]]; do
	case $1 in
		-h|--help)
			echo "Please read the README.md file for usage information."
			exit
			;;
		-f|--font)
			FONT="$2"
			shift # past argument
			shift # past value
			;;
		-l|--logo)
			LOGO=$2
			shift # past argument
			shift # past value
			;;
		-t|--template)
			TEMPLATE=$2
			shift # past argument
			shift # past value
			;;
		-*|--*)
			echo "Unknown option $1"
			exit 1
			;;
		*)
			POSITIONAL_ARGS+=("$1") # save positional arg
			shift # past argument
			;;
	esac
done

LOGO_BASENAME=$(basename "$LOGO")

set -- "${POSITIONAL_ARGS[@]}" # restore positional parameters

# Remove temp-folder from previous run and create new one
rm -rf .tmp
mkdir -p .tmp

# Copy necessary files
cp "$TEMPLATE" .tmp/cover.tex
cp "$LOGO" .tmp

cd .tmp

# Turn every \ into \\
ESCAPED_TITLE=${1//\\/\\\\}
ESCAPED_PRE=${2//\\/\\\\}
ESCAPED_PST=${3//\\/\\\\}
ESCAPED_FONT=${FONT//\\/\\\\}
ESCAPED_LOGO=${LOGO_BASENAME//\\/\\\\}
ESCAPED_LOGO=${ESCAPED_LOGO//\//\\/} # Turn / into \/ for sed

# Replace the strings in the copy of the LaTeX document
sed -i "s/((TITLE))/$ESCAPED_TITLE/" cover.tex
sed -i "s/((PRE))/$ESCAPED_PRE/" cover.tex
sed -i "s/((POST))/$ESCAPED_POST/" cover.tex
sed -i "s/%((FONT))/\\\\usepackage{$ESCAPED_FONT}/" cover.tex
sed -i "s/.\/logo/.\/$ESCAPED_LOGO/" cover.tex

pdflatex ./cover.tex
convert -background white -alpha remove cover.pdf cover.png

mv cover.png ../
cd ../

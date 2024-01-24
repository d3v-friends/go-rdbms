#!/bin/bash

# local tag
TAG=$1
HAS=$(git tag --list | grep $TAG)

if ! [ -z "$HAS" ]; then
	git tag -d $TAG  || { echo "❌fail delete tag"; exit 1; }
else
	echo "❌not found local tag: tag=$TAG"
fi

# remote tag
ORIGIN=$2
if [ -z "$ORIGIN" ]; then
	ORIGIN=origin
fi

HAS=$(git ls-remote tags $ORIGIN | grep "$TAG")
if ! [ -z "$HAS" ]; then
	git push -d $ORIGIN $TAG || { echo "❌fail delete tag"; exit 1; }
else
	echo "❌not found remote tag: origin=$ORIGIN, tag=$TAG"
fi
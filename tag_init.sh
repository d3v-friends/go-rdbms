#!/bin/bash

ORIGIN=$1

if [ -z "$ORIGIN" ]; then
	ORIGIN="origin"
fi

git fetch
git push -d "$ORIGIN" $(git tag -l)
git tag -d $(git tag --list)
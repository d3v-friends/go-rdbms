#!/bin/bash

# ex) sh tag.sh 1.0.0
TZ=utc

VERSION=$1
if [ -z "$VERSION" ]; then
    VERSION="1.0.0"
fi
DATE=$(date '+%Y%m%d%H%M%S')

ORIGIN=$2
if [ -z "$ORIGIN" ]; then
	ORIGIN="origin"
fi

TAG="v$VERSION-$DATE"

# delete tag
git tag -d "$TAG"
git push -d "$ORIGIN" "$TAG"

# create tag
git tag "$TAG"
git push --tags "$ORIGIN" "$TAG"

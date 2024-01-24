#!/bin/bash

# ex) sh tag.sh 1.0.0 origin
TZ=utc
VERSION=$1
if [ -z "$VERSION" ]; then
    echo "no has version"
    exit 1;
fi

ORIGIN=$2
if [ -z "$ORIGIN" ]; then
	ORIGIN="origin"
fi

TAG="v$VERSION"

# delete tag
git tag -d "$TAG"
git push -d "$ORIGIN" "$TAG"

# create tag
git tag "$TAG"
git push --tags "$ORIGIN" "$TAG"

# done

echo "💯 done"

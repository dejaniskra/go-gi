#!/bin/bash

# Check if a version argument is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <version>"
  exit 1
fi

VERSION="$1"

# Run the commands
git tag "v$VERSION"
git push origin "v$VERSION"
GOPROXY=proxy.golang.org go list -m "github.com/dejaniskra/go-gi@v$VERSION"

#!/usr/bin/env bash

VERSION=$1

echo "// Code generated. DO NOT EDIT.

package config

const Version = \"v$VERSION\"" > config/version.go

go fmt config/version.go

git add config/version.go
git commit -m "bump: $VERSION"
git tag "v${VERSION}" -m "v${VERSION}"

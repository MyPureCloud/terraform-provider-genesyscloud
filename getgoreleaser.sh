#!/bin/sh
set -e

if test "$DISTRIBUTION" = "pro"; then
	echo "Using Pro distribution..."
	RELEASES_URL="https://github.com/goreleaser/goreleaser-pro/releases"
	FILE_BASENAME="goreleaser-pro"
	LATEST="$(curl -sf https://goreleaser.com/static/latest-pro)"
else
	echo "Using the OSS distribution..."
	RELEASES_URL="https://github.com/goreleaser/goreleaser/releases"
	FILE_BASENAME="goreleaser"
	LATEST="$(curl -sf https://goreleaser.com/static/latest)"
fi

test -z "$VERSION" && VERSION="$LATEST"

test -z "$VERSION" && {
	echo "Unable to get goreleaser version." >&2
	exit 1
}

test -z "$TMPDIR" && TMPDIR="$(mktemp -d)"
export TAR_FILE="$TMPDIR/${FILE_BASENAME}_$(uname -s)_$(uname -m).tar.gz"

(
	cd "$TMPDIR"
	echo "Downloading GoReleaser $VERSION..."
	curl -sfLo "$TAR_FILE" \
		"$RELEASES_URL/download/$VERSION/${FILE_BASENAME}_$(uname -s)_$(uname -m).tar.gz"

)

tar -xf "$TAR_FILE" -C "$TMPDIR"
"${TMPDIR}/goreleaser" "$@"

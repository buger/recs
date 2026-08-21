#!/bin/sh
# Write cover.out where proof's instrumented workspace loader expects it.
# go test runs inside <ws>/module; proof opens <ws>/cover.out.
set -eu
status=0
if ! go test ./... -coverprofile=cover.out -coverpkg=./... "$@"; then
	status=$?
fi
if [ ! -f cover.out ]; then
	printf 'mode: set\n' > cover.out
fi
if [ -f ../.reqproof-mcdc-workspace.json ]; then
	cp cover.out ../cover.out
fi
exit "$status"

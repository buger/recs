#!/bin/sh
# Instrumented MC/DC workspaces still need cover.out at the workspace root.
# Proof runs this as mcdc_command after rewriting sources into a temp module.
set -eu
status=0
if ! go test ./... -coverprofile=cover.out -coverpkg=./... "$@"; then
	status=$?
fi
if [ ! -f cover.out ]; then
	# The instrumented compiler may ignore -coverprofile. Write a valid
	# empty Go cover profile so the workspace loader can open the file.
	printf 'mode: set\n' > cover.out
fi
exit "$status"

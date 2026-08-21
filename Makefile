# File-native CRM
.PHONY: all test ape hello clean

all:
	go build -o crm ./cmd/crm

test:
	go test ./...

# Cosmopolitan APE: wrapper + embedded host Go blob.
# Requires cosmocc on PATH or in $$HOME/.local/cosmocc/bin.
ape:
	./scripts/build-ape.sh

hello:
	./scripts/build-ape.sh

clean:
	rm -rf crm crm.com .build

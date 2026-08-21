# File-native records
.PHONY: all test clean

all:
	go build -o recs ./cmd/recs

test:
	go test ./...

clean:
	rm -rf recs crm

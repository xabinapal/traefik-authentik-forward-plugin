.PHONY: lint test vendor clean

PACKAGES := $(shell go list ./...)

default: lint test

lint:
	golangci-lint run

test:
	go test -v -cover ./...

yaegi_test:
	$(foreach pkg, $(PACKAGES), yaegi test -v $(pkg);)
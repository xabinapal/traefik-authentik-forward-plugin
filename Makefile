.POSIX:

$(V).SILENT:

.PHONY: vendor clean lint test yaegi_test e2e_test

PACKAGES := $(shell go list ./...)

default: lint test

vendor:
	go mod vendor

clean:
	rm -rf ./vendor

lint:
	golangci-lint run

test:
	go test -v -cover ./...

yaegi_test:
	$(foreach pkg, $(PACKAGES), yaegi test -v $(pkg);)

e2e_test:
	cd e2e && npm install --include=dev && npm run test

all: clean check-quality build test golangci

ALL_PACKAGES=$(shell go list ./...)
SOURCE_DIRS=$(shell go list ./... | cut -d "/" -f2 | uniq)

clean:
	rm -rf ./out
	GO111MODULE=on go mod tidy -v

check-quality: setup lint fmt imports vet

setup:
	GO111MODULE=off go get -v golang.org/x/tools/cmd/goimports
	GO111MODULE=off go get -v golang.org/x/lint/golint

lint:
	@if [[ `golint $(ALL_PACKAGES) | { grep -vwE "exported (var|function|method|type|const) \S+ should have comment" || true; } | wc -l | tr -d ' '` -ne 0 ]]; then \
          golint $(ALL_PACKAGES) | { grep -vwE "exported (var|function|method|type|const) \S+ should have comment" || true; }; \
          exit 2; \
    fi;

fmt:
	gofmt -l -s -w $(SOURCE_DIRS)

imports:
	./scripts/lint.sh check_imports

vet:
	go vet ./...

cyclo:
	gocyclo -over 6 $(SOURCE_DIRS)


fix_imports:
	goimports -l -w .

golangci:
	GO111MODULE=off go get -v github.com/golangci/golangci-lint/cmd/golangci-lint
	golangci-lint run -v --deadline 5m0s


build:
	@echo "Building './out/albatross'"
	@mkdir -p ./out
	@go build -o out/albatross ./cmd/service

test:
	go test -race ./...

testcodecov:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

run: build
	./out/albatross

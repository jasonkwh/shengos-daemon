.PHONY: lint test-unit test-clean-cache test-coverage serve build build-amd64 build-arm64 build-mocks update
.DEFAULT_GOAL := serve

update:
	nix flake update

lint:
	golangci-lint run -c .golangci.yml

test-unit:
	go test -count=1 ./...

test-clean-cache:
	go clean -testcache

test-coverage:
	go test -count=1 -covermode=atomic -coverprofile=coverage.out $$(go list ./... | grep -vE '/(proto|mocks|graph|cmd)(/|$$)')

serve:
	go run main.go --config=./config/config.yaml serve

build: build-amd64 build-arm64

build-amd64:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/shengosd_amd64 .

build-arm64:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o dist/shengosd_arm64 .

build-mocks:
	cd mocks/ && rm -rf -- */ && mockery

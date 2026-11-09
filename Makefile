.PHONY: default lint test-unit test-clean-cache serve build-proto

default: serve

lint:
	golangci-lint run -c .golangci.yml

test-unit:
	go test -count=1 ./...

test-clean-cache:
	go clean -testcache

test-coverage:
	go test -count=1 -covermode=atomic -coverprofile=coverage.out $$(go list ./... | grep -vE '/(proto|mocks|graph)(/|$$)')

serve:
	go run main.go --config=./config/config.yaml serve

build-proto:
	protoc \
		--go_out=. --go-grpc_out=. \
		--go_opt=paths=source_relative --go-grpc_opt=paths=source_relative \
		./proto/daemon_svc.proto

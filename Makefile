.PHONY: lint test-unit test-clean-cache test-coverage serve build build-amd64 build-arm64 build-mocks update update-vendor-hash
.DEFAULT_GOAL := serve

update:
	nix flake update

update-vendor-hash:
	@set -eu; \
	backup=$$(mktemp); log=$$(mktemp); \
	cp flake.nix "$$backup"; \
	trap 'rm -f "$$backup" "$$log"' EXIT; \
	count=$$(grep -c '^[[:space:]]*vendorHash =' flake.nix || true); \
	if [ "$$count" -ne 1 ]; then echo "expected one vendorHash in flake.nix, found $$count" >&2; exit 1; fi; \
	sed -Ei 's|^([[:space:]]*vendorHash = ).*;|\1nixpkgs.lib.fakeHash;|' flake.nix; \
	if nix build --no-link .#packages.x86_64-linux.shengos-daemon >"$$log" 2>&1; then \
	  cat "$$log"; cp "$$backup" flake.nix; echo "fakeHash build unexpectedly succeeded" >&2; exit 1; \
	fi; \
	cat "$$log"; \
	hash=$$(sed -nE 's/^[[:space:]]*got:[[:space:]]*(sha256-[A-Za-z0-9+/=]+).*$$/\1/p' "$$log" | tail -n 1); \
	if ! printf '%s\n' "$$hash" | grep -Eq '^sha256-[A-Za-z0-9+/=]+$$'; then \
	  cp "$$backup" flake.nix; echo "could not extract vendorHash from Nix output; restored flake.nix" >&2; exit 1; \
	fi; \
	sed -Ei "s|^([[:space:]]*vendorHash = )nixpkgs\\.lib\\.fakeHash;|\\1\"$$hash\";|" flake.nix; \
	echo "vendorHash set to $$hash; verifying package build"; \
	nix build --no-link .#packages.x86_64-linux.shengos-daemon

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

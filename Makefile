PREFIX=/usr/local

.PHONY: test update vendor fmt lint vet sloc clean install run run-dev hotreload act eslint run-cli

filesender: cmd/filesender/main.go
	go build $(GOBUILDFLAGS) -o $@ codeberg.org/filesender/filesender-next/cmd/filesender

test:
	go test -v ./...

update:
	# update Go dependencies
	go get -t -u ./...
	go mod tidy

vendor:
	go mod vendor

fmt:
	gofumpt -w . || go fmt ./...

lint:
	golangci-lint run -E staticcheck,revive,gocritic --timeout=5m

vet:
	go vet ./...

sloc:
	tokei . || cloc .

clean:
	rm -f filesender
	rm -rf vendor

install: filesender
	install -D filesender $(DESTDIR)$(PREFIX)/bin/filesender

run:
	mkdir -p ./data
	STATE_DIRECTORY=./data go run ./cmd/filesender

run-dev:
	mkdir -p ./data
	FILESENDER_AUTH_METHOD=dummy STATE_DIRECTORY=./data go run ./cmd/filesender

hotreload:
	watchexec --shell=none -r -w ./internal -- make run-dev

act:
	act --container-architecture linux/amd64 --workflows .forgejo/workflows/tests.yaml

eslint:
	npx eslint

run-cli:
	go run ./cmd/filesender-cli $(ARGS)

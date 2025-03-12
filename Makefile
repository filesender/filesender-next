PREFIX=/usr/local

.PHONY: test update vendor fmt lint vet sloc clean install run

filesender: cmd/filesender/main.go
	go build $(GOBUILDFLAGS) -o $@ codeberg.org/filesender/filesender-next/cmd/filesender

test:
	go test ./...

update:
	# update Go dependencies
	go get -t -u ./...
	go mod tidy

vendor:
	go mod vendor

fmt:
	gofumpt -w . || go fmt ./...

lint:
	golangci-lint run -E stylecheck,revive,gocritic --timeout=5m

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
	FILESENDER_DEBUG=true STATE_DIRECTORY=./data go run ./cmd/filesender

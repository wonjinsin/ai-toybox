BIN_DIR := bin
APP := whisper-local

.PHONY: build test check clean

build:
	go build -o $(BIN_DIR)/$(APP) ./cmd/$(APP)

test:
	go test -race -count=1 -coverpkg=./... -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | awk '/^total:/ { print; if ($$3 + 0 < 80) exit 1 }'

check:
	test -z "$$(gofmt -l cmd internal)"
	go vet ./...

clean:
	rm -rf $(BIN_DIR)

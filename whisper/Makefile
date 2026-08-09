BIN_DIR := bin
APP := whisper-local

.PHONY: build clean

build:
	go build -o $(BIN_DIR)/$(APP) ./cmd/$(APP)

clean:
	rm -rf $(BIN_DIR)

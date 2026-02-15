BIN_DIR := bin
SHARED_DIR := shared

SHARED_DEPS := $(SHARED_DIR)/consts.go
CLIENT_DEPS := $(wildcard client/*.go) $(wildcard client/ansi/*.go) $(wildcard client/input/*.go) $(wildcard client/buffer/*.go)

build: $(BIN_DIR)/dungeon-client $(BIN_DIR)/dungeon-server

$(BIN_DIR)/dungeon-client: $(CLIENT_DEPS) $(SHARED_DEPS)
	go build -o $@ ./client

$(BIN_DIR)/dungeon-server: server/main.go $(SHARED_DEPS)
	go build -o $@ ./server

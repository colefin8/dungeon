BIN_DIR := bin
SHARED_DIR := shared

SHARED_DEPS := $(SHARED_DIR)/consts.go
CLIENT_DEPS := $(wildcard client/*.go client/**/*.go)
SERVER_DEPS := $(wildcard server/*.go server/**/*.go)

build: $(BIN_DIR)/dungeon-client $(BIN_DIR)/dungeon-server

$(BIN_DIR)/dungeon-client: $(CLIENT_DEPS) $(SHARED_DEPS)
	go build -o $@ ./client

$(BIN_DIR)/dungeon-server: $(SERVER_DEPS) $(SHARED_DEPS)
	go build -o $@ ./server

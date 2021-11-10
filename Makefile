.PHONY: all test clean serve serve-tcp

OBJ := bitty-bingo
BUILD_DIR := build
GO_ARGS :=
SERVE_ARGS := $(shell grep -s -v "^\#" .env)

all: $(BUILD_DIR)/$(OBJ)

test:
	go test ./... --cover

clean:
	rm -rf $(BUILD_DIR)

serve: $(BUILD_DIR)/$(SERVER_OBJ)
	$(SERVE_ARGS) $<

serve-tcp: $(BUILD_DIR)/$(SERVER_OBJ)
	sudo setcap cap_net_bind_service=+ep $<
	$(SERVE_ARGS) HTTP_PORT=80 HTTPS_PORT=443 $<

$(BUILD_DIR):
	mkdir -p $@

$(BUILD_DIR)/$(OBJ): test | $(BUILD_DIR)
	$(GO_ARGS) go build -o $@ \
		github.com/jacobpatterson1549/bitty-bingo/cmd/server

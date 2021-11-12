.PHONY: all test clean serve

OBJ := bitty-bingo
BUILD_DIR := build
GO_ARGS :=
SERVE_ARGS := $(shell grep -s -v "^\#" .env)

all: $(BUILD_DIR)/$(OBJ)

test:
	go test ./... --cover

clean:
	rm -rf $(BUILD_DIR)

serve: $(BUILD_DIR)/$(OBJ)
	$(SERVE_ARGS) $<

$(BUILD_DIR):
	mkdir -p $@

$(BUILD_DIR)/$(OBJ): test | $(BUILD_DIR)
	$(GO_ARGS) go build -o $@ \
		github.com/jacobpatterson1549/bitty-bingo/cmd/server

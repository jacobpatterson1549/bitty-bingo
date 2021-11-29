.PHONY: all test coverage clean serve

OBJ := bitty-bingo
BUILD_DIR := build
COVERAGE_OBJ := .coverage.out
GO_ARGS :=
SERVE_ARGS := $(shell grep -s -v "^\#" .env)
SRC := $(shell find bingo/ cmd/ server/ go.mod go.sum)
GO := $(GO_ARGS) go
GO_TEST := $(GO) test ./...
GO_BUILD := $(GO) build
GO_TOOL := $(GO) tool

all: $(BUILD_DIR)/$(OBJ)

test: $(SRC)
	$(GO_TEST) --cover

coverage: $(COVERAGE_OBJ)
	$(GO_TOOL) cover -html=$<

clean:
	rm -rf $(BUILD_DIR) $(COVERAGE_OBJ)

serve: $(BUILD_DIR)/$(OBJ)
	$(SERVE_ARGS) $<

$(BUILD_DIR):
	mkdir -p $@

$(BUILD_DIR)/$(OBJ): test | $(BUILD_DIR)
	$(GO_BUILD) -o $@ \
		github.com/jacobpatterson1549/bitty-bingo/cmd/server

$(COVERAGE_OBJ): $(SRC)
	$(GO_TEST) -coverprofile=$@
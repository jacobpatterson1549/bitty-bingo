.PHONY: all test coverage clean serve

OBJ := bitty-bingo
BUILD_DIR := build
COVERAGE_OBJ := coverage.out

GO_ARGS :=
GO := $(GO_ARGS) go
GO_TOOL := $(GO) tool
GO_TEST := $(GO) test ./...
GO_BUILD := $(GO) build

all: $(BUILD_DIR)/$(OBJ)

test: $(BUILD_DIR)/$(COVERAGE_OBJ)

coverage: $(BUILD_DIR)/$(COVERAGE_OBJ)
	$(GO_TOOL) cover -html=$<

clean:
	rm -rf $(BUILD_DIR)

serve: $(BUILD_DIR)/$(OBJ)
	$(shell grep -s -v "^\#" .env) $<

$(BUILD_DIR):
	mkdir -p $@

$(BUILD_DIR)/$(OBJ): $(BUILD_DIR)/$(COVERAGE_OBJ) | $(BUILD_DIR)
	$(GO_BUILD) -o $@

$(BUILD_DIR)/$(COVERAGE_OBJ): | $(BUILD_DIR)
	$(GO_TEST) -coverprofile=$@
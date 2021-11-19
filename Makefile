.PHONY: all test coverage clean serve

OBJ := bitty-bingo
BUILD_DIR := build
COVERAGE_OBJ := .coverage.out
GO_ARGS :=
SERVE_ARGS := $(shell grep -s -v "^\#" .env)
SRC := $(shell find bingo/ cmd/ server/ go.mod go.sum)

all: $(BUILD_DIR)/$(OBJ)

test: $(SRC)
	go test ./... --cover

coverage: $(COVERAGE_OBJ)
	go tool cover -html=$<

clean:
	rm -rf $(BUILD_DIR) $(COVERAGE_OBJ)

serve: $(BUILD_DIR)/$(OBJ)
	$(SERVE_ARGS) $<

$(BUILD_DIR):
	mkdir -p $@

$(BUILD_DIR)/$(OBJ): test | $(BUILD_DIR)
	$(GO_ARGS) go build -o $@ \
		github.com/jacobpatterson1549/bitty-bingo/cmd/server

$(COVERAGE_OBJ): $(SRC)
	go test -cover ./... -coverprofile=$@
.PHONY: all test clean

OBJ       := bitty-bingo
BUILD_DIR := build
GO_ARGS   :=

all: $(BUILD_DIR)/$(OBJ)

test:
	go test ./... --cover

clean:
	rm -rf $(BUILD_DIR)

$(BUILD_DIR):
	mkdir -p $@

$(BUILD_DIR)/$(OBJ): test | $(BUILD_DIR)
	$(GO_ARGS) go build -o $@ \
		github.com/jacobpatterson1549/bitty-bingo/cmd/server

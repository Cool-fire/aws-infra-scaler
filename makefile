.PHONY: all
all: build

GOENVS ?= CGO_ENABLED=0
GO ?= $(GOENVS) go
LDFLAGS ?= -ldflags "-s -w"
GOBUILD ?= $(GO) build $(LDFLAGS)
BIN_DIR ?= bin
TARGET ?= $(BIN_DIR)/scaler

.PHONY: build
build:
	$(GOBUILD) -o $(TARGET) ./

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)

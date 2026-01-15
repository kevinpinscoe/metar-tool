BIN       := metar-tool
PKG       := .
BUILD_DIR := bin

# You can override these:make install PREFIX=$HOME/.local
PREFIX    ?= /usr/local
BINDIR    ?= $(PREFIX)/bin

GO        ?= go
GOFLAGS   ?=
LDFLAGS   ?=

.PHONY: all build clean run test fmt vet install uninstall

all: build

build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BIN) $(PKG)

run: build
	./$(BUILD_DIR)/$(BIN) --forecast nws kmrx
	./$(BUILD_DIR)/$(BIN) --obs ktys
	./$(BUILD_DIR)/$(BIN) --obs ktys --json


fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

clean:
	rm -rf $(BUILD_DIR)

install: build
	@install -d "$(BINDIR)"
	install -m 0755 "$(BUILD_DIR)/$(BIN)" "$(BINDIR)/$(BIN)"
	@echo "Installed $(BIN) to $(BINDIR)/$(BIN)"

uninstall:
	rm -f "$(BINDIR)/$(BIN)"
	@echo "Removed $(BINDIR)/$(BIN)"

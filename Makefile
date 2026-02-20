BIN       := metar-tool
PKG       := .
BUILD_DIR := bin

# override: make install PREFIX=/usr/local
PREFIX    ?= $(HOME)/.local
BINDIR    ?= $(PREFIX)/bin

GO        ?= go
GOFLAGS   ?=
VERSION   ?= 2.0.5
LDFLAGS   ?= -X 'main.Version=$(VERSION)'

.PHONY: all build clean run run-decode test fmt vet lint install uninstall

all: build

build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BIN) $(PKG)

# Smoke tests
run: build
	@echo "== nws afd =="
	./$(BUILD_DIR)/$(BIN) --forecast nws kmrx
	@echo
	@echo "== obs raw =="
	./$(BUILD_DIR)/$(BIN) --obs ktys
	@echo
	@echo "== obs raw -> file =="
	@rm -f ktys
	./$(BUILD_DIR)/$(BIN) --obs ktys --output ktys
	@echo "wrote ./ktys"
	@echo
	@echo "== obs json =="
	./$(BUILD_DIR)/$(BIN) --obs ktys --json
	@echo
	@echo "== obs json pretty =="
	./$(BUILD_DIR)/$(BIN) --obs ktys --json --pretty
	@echo
	@echo "== decode from raw pipe =="
	./$(BUILD_DIR)/$(BIN) --obs ktys | ./$(BUILD_DIR)/$(BIN) --decode
	@echo
	@echo "== decode from json pipe =="
	./$(BUILD_DIR)/$(BIN) --obs ktys --json | ./$(BUILD_DIR)/$(BIN) --decode
	@echo "== version =="
	./$(BUILD_DIR)/$(BIN) --version

# Decode-related checks
run-decode: build
	./$(BUILD_DIR)/$(BIN) --obs ktys | ./$(BUILD_DIR)/$(BIN) --decode
	./$(BUILD_DIR)/$(BIN) --obs ktys --json | ./$(BUILD_DIR)/$(BIN) --decode

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

lint: fmt vet test

clean:
	rm -rf $(BUILD_DIR)
	rm -f ktys

install: build
	@install -d "$(BINDIR)"
	install -m 0755 "$(BUILD_DIR)/$(BIN)" "$(BINDIR)/$(BIN)"
	@echo "Installed $(BIN) to $(BINDIR)/$(BIN)"

uninstall:
	rm -f "$(BINDIR)/$(BIN)"
	@echo "Removed $(BINDIR)/$(BIN)"
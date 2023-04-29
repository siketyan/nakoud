GO = go

bin/nakoud-proxy:
	$(GO) build -o ./bin/nakoud-proxy ./cmd/proxy

.PHONY: tools
tools:
	$(MAKE) -C tools

.PHONY: all
.DEFAULT_GOAL := build
build: bin/nakoud-proxy

.PHONY: all
all: tools build

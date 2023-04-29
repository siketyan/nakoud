export GOBIN ?= $(shell pwd)/bin

GCI = $(GOBIN)/gci
GOFUMPT = $(GOBIN)/gofumpt

$(GCI):
	go install 'github.com/daixiang0/gci'

$(GOFUMPT):
	go install 'mvdan.cc/gofumpt'

.PHONY: all
.DEFAULT_GOAL := all
all: $(GCI) $(GOFUMPT)

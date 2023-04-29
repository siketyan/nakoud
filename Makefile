.PHONY: tools
tools:
	$(MAKE) -C tools

.PHONY: all
.DEFAULT_GOAL := all
all: tools

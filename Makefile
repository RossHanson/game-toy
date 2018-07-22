# note: call scripts from /scripts

PACKAGE = game-toy
DATE ?= $(shell date +%FT%T%z)
GOPATH = $(CURDIR)/.gopath
BASE = $(GOPATH)/src
BIN = $(GOPATH)/bin
GO = go
GODOC = godoc
GOFMT = gofmt
GLIDE = glide

$(BASE):
	@mkdir -p $(dir $@)
	@ln -sf $(CURDIR)/src $@

.PHONY: all
all: | $(BASE)
	cd $(BASE) && $(GO) build -o $(CURDIR)/bin/$(PACKAGE) main.go

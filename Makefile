# note: call scripts from /scripts

PACKAGE = game-toy
DATE ?= $(shell date +%FT%T%z)
export GOPATH = $(CURDIR)/.gopath
BASE = $(GOPATH)/src
BIN = $(GOPATH)/bin
GO = go
GODOC = godoc
GOFMT = gofmt
PKGS     = $(or $(PKG),$(shell cd $(BASE) && env GOPATH=$(GOPATH) $(GO) list ./... | grep -v "_"))
TESTPKGS = $(shell env GOPATH=$(GOPATH) $(GO) list -f '{{ if or .TestGoFiles .XTestGoFiles }}{{ .ImportPath }}{{ end }}' $(PKGS))
TIMEOUT =30s

Q = $(if $(filter 1,$V),,@)

$(BASE):
	@mkdir -p $(dir $@)
	@ln -sf $(CURDIR)/src $@

.PHONY: all
all: | $(BASE)
	cd $(BASE) && $(GO) build -o $(CURDIR)/bin/$(PACKAGE) main.go

.PHONY: ($TEST_TARGETS) check test tests

$(TEST_TARGETS): NAME=$(MAKECMDGOALS:test-%=%)
$(TEST_TARGETS): test
check test tests: | $(BASE) ; $(info running $(NAME=%=% )tests...) @
	$Q cd $(BASE) && $(GO) test -timeout $(TIMEOUT) $(TESTPKGS)

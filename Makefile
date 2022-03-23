SRC := $(shell find . -name *.go)
BIN := gitup

.PHONY: all clean test build install upgrade help

all: 			# default action
	@pre-commit install --install-hooks
	@git config commit.template .git-commit-template

clean:			# clean-up environment
	@find . -name '*.sw[po]' -delete
	rm -f $(BIN)

test:			# run test
	gofmt -w -s $(SRC)
	go test -cover -failfast -timeout 2s ./...

build: $(BIN)	# build the binary/library

install: $(BIN)	# install the binary to local env
	go install ./...

upgrade:		# upgrade all the necessary packages
	pre-commit autoupdate

help:			# show this message
	@printf "Usage: make [OPTION]\n"
	@printf "\n"
	@perl -nle 'print $$& if m{^[\w-]+:.*?#.*$$}' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?#"} {printf "    %-18s %s\n", $$1, $$2}'

$(BIN): test

$(BIN): $(SRC)
	@go mod tidy
	go build -o $@ $(SRC)

.PHONY: build
build:
		go build -o rstat -v ./cmd/utility

.DEFAULT_GOAL := build 
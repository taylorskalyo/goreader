BUILD_DIR := bin

default: all

all: test build

bin/goreader:
	go build -o bin/goreader

.PHONY: build
build: bin/goreader

.PHONY: test
test:
	go test -cover -coverprofile=coverage.txt -covermode=atomic -coverpkg=./config,./epub,./render,./state,./views ./...

.PHONY: cover
cover: test
	go tool cover -html=coverage.txt

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	rm coverage.txt

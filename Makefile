.PHONY: build run test docker-build docker-run k8s-deploy clean generate

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOGET=$(GOCMD) get
BINARY_NAME=eth-balance-proxy
MAIN_DIR=cmd/server

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./$(MAIN_DIR)

run: build
	./$(BINARY_NAME)

test:
	$(GOTEST) -v ./...

generate:
	$(GOCMD) generate ./...

deps:
	$(GOMOD) download

tidy:
	$(GOMOD) tidy

docker-build:
	docker build -t $(BINARY_NAME):latest .

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
.PHONY: test lint fmt vet build

test:
	go test ./... -v -count=1

cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint: vet
	@echo "Lint passed"

fmt:
	gofmt -w .

vet:
	go vet ./...

build:
	go build ./...

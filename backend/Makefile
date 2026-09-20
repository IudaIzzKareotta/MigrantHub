.PHONY: run test build fmt vet tidy

run:
	go run ./services/api/cmd/server

test:
	go test ./...

build:
	go build -o bin/api ./services/api/cmd/server

fmt:
	gofmt -w .

vet:
	go vet ./...

tidy:
	go mod tidy
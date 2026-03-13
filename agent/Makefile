# Usage: make run ARGS="start --pg-host localhost"
run:
	go run cmd/main.go $(ARGS)

build:
	CGO_ENABLED=0 go build -ldflags "-X main.Version=$(VERSION)" -o databasus-agent ./cmd/main.go

test:
	go test -count=1 -failfast ./internal/...

lint:
	golangci-lint fmt ./cmd/... ./internal/... && golangci-lint run ./cmd/... ./internal/...

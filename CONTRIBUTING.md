# Contributing

Thanks for your interest in contributing to the Sonzai Go SDK.

## Setup

```bash
git clone https://github.com/sonz-ai/sonzai-go.git
cd sonzai-go
go test ./...
```

## Running Tests

```bash
go test ./... -v          # All tests
go test ./... -cover      # With coverage
go vet ./...              # Static analysis
```

## Guidelines

- Keep zero external dependencies. Use only the Go standard library.
- All public functions must have doc comments.
- Add tests for new functionality.
- Run `go vet` and `gofmt` before submitting.

## Pull Requests

1. Fork the repo and create your branch from `main`.
2. Add tests for any new or changed behavior.
3. Ensure all tests pass.
4. Open a pull request with a clear description.

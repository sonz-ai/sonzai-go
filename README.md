# Sonzai Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/sonz-ai/sonzai-go.svg)](https://pkg.go.dev/github.com/sonz-ai/sonzai-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/sonz-ai/sonzai-go)](https://goreportcard.com/report/github.com/sonz-ai/sonzai-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

The official Go SDK for the [Sonzai Character Engine API](https://sonz.ai). Build AI characters with persistent memory, evolving personality, and real-time voice.

Zero dependencies. Uses only the Go standard library.

## Documentation

Full API documentation is available at [sonz.ai/docs](https://sonz.ai/docs).

See the [api reference](api.md) for a complete list of methods and types.

## Installation

```go
import (
    sonzai "github.com/sonz-ai/sonzai-go" // imported as sonzai
)
```

```bash
go get github.com/sonz-ai/sonzai-go@v1.12.0
```

## Getting Started

```go
package main

import (
    "context"
    "fmt"

    sonzai "github.com/sonz-ai/sonzai-go"
)

func main() {
    client := sonzai.NewClient("my-api-key") // defaults to os.LookupEnv("SONZAI_API_KEY")

    // Stream a chat response
    err := client.Agents.ChatStream(
        context.Background(),
        "agent-id",
        sonzai.ChatOptions{
            Messages: []sonzai.ChatMessage{{Role: "user", Content: "Hello!"}},
            UserID:   "user-123",
        },
        func(event sonzai.ChatStreamEvent) error {
            fmt.Print(event.Content())
            return nil
        },
    )
    if err != nil {
        panic(err)
    }
}
```

See the [examples](examples/) directory for more.

## Requirements

This library requires Go 1.22 or later.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT - see [LICENSE](LICENSE) for details.

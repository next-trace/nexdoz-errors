<p align="center"><img src="art/nexdoz.webp" alt="Diabuddy Error package"></p>

## Introduction

Diabuddy Errors provides a shared library for the nexdoz platform that standardises the structure of API errors across every Go service.

### Install

With Go's module support, `go [build|run|test]` fetches the module automatically once it is imported:
```go
import "github.com/next-trace/nexdoz-errors"
```

Or pull it explicitly:
```bash
go get github.com/next-trace/nexdoz-errors
```

### Use

```go
package main

import (
    "encoding/json"

    nexdozErrors "github.com/next-trace/nexdoz-errors"
)

type User struct{}

func main() {
    data := []byte("{response from a specific api}")
    payload := &struct {
        User   *User  `json:"user"`
        Gender string `json:"gender"`
    }{}

    if err := json.Unmarshal(data, &payload); err != nil {
        _ = nexdozErrors.NewApiError(
            nexdozErrors.UnprocessableEntityErrorType,
            "unprocessable response data",
            nexdozErrors.WithInternalError(err),
        )
    }
}
```

Returned errors satisfy `errors.Is`/`errors.As` because `ApiError.Unwrap()` exposes the wrapped internal cause, so upstream code can walk the chain without losing context.

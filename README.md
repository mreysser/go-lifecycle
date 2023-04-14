# go-lifecycle

Dirt simple library to manage graceful shutdown of a multi-goroutine application upon receiving SIGTERM or any other user-configured signal(s). This is useful for stopping stateful workloads and releasing resources.

Under the hood, it does the following:

1. Creates a goroutine to listen for the specified signal(s)

1. Upon receiving one of the signals, it executes any registered shutdown handlers

1. After all shutdown handlers have been completed, it calls `cancel()` on the specified context

All functions are thread safe.

# Example - Echo Server

```go
package main

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	lifecycle "github.com/mreysser/go-lifecycle"
	log "github.com/sirupsen/logrus"
)

func main() {
	token := lifecycle.GetDefaultLifecycleToken()

	e := echo.New()
	e.GET("/", func(c echo.Context) error { return c.String(http.StatusOK, "Hello, world!") })

	go func() {
		if err := e.StartServer(&http.Server{Addr: ":8080"}); err != nil && err != http.ErrServerClosed {
			log.Errorf("server failed to start: %s", err.Error())
			token.TerminateLifecycle()
		}
	}()

	token.RegisterShutdownHandler(func(ctx context.Context) { e.Shutdown(ctx) })
	<-token.GetContext().Done()
	log.Warn("application exit")
}

```
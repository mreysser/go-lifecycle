package lifecycle

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Interface for common lifecycle management tasks, provided for convenient dependency injection. A
// mock is also provided in the mocks folder.
//
//go:generate mockery --name LifecycleManager --structname MockLifecycleManager
type LifecycleManager interface {
	RegisterShutdownHandler(handler ShutdownHandler)
	TerminateLifecycle()
	GetContext() context.Context
}

// The LifecycleToken is a singleton that holds all lifecycle-related state for the application.
type LifecycleToken struct {
	ctx         context.Context
	cancel      context.CancelFunc
	signalChan  chan os.Signal
	handlerList []ShutdownHandler
	alive       bool
	lock        sync.Mutex
}

// ShutdownHandler is a callback function. Goroutines can register a handler with the
// [LifecycleToken] to be notified when the application is being terminated.
type ShutdownHandler func(ctx context.Context)

var token *LifecycleToken
var lifecycleLock sync.Mutex

// Builds the default [LifecycleToken] singleton, if not already built, and then returns it.
//
// Uses context.Background() for default context, and SIGTERM as the default termination signal.
//
// After returning, the [LifecycleToken] will be alive and listening for SIGTERM.
func GetDefaultLifecycleToken() *LifecycleToken {
	return GetLifecycleToken(context.Background(), []syscall.Signal{syscall.SIGTERM})
}

// Builds the [LifecycleToken] singleton, if not already built, and then returns it.
//
// Uses baseContext to build the context for the lifecycle.
//
// Uses terminationSignals to register with the system for the specified signal(s). If no signals
// are provided, the only way to stop the lifecycle is via [LifecycleToken.TerminateLifecycle]
func GetLifecycleToken(baseContext context.Context, terminationSignals []syscall.Signal) *LifecycleToken {
	lifecycleLock.Lock()
	defer lifecycleLock.Unlock()

	if token != nil {
		return token
	}

	ctx, cancel := context.WithCancel(baseContext)
	signalChan := make(chan os.Signal, 1)

	if len(terminationSignals) > 0 {
		for _, sig := range terminationSignals {
			signal.Notify(signalChan, sig)
		}
	}

	token = &LifecycleToken{
		ctx:         ctx,
		cancel:      cancel,
		signalChan:  signalChan,
		handlerList: make([]ShutdownHandler, 0),
		alive:       true,
	}

	go token.blockUntilTerminationSignal()

	return token
}

// Registers a [ShutdownHandler] to be executed when the lifecycle is terminated.
//
// If the lifecycle is already terminated, the handler will be executed immediately in a
// synchronous fashion.
func (t *LifecycleToken) RegisterShutdownHandler(handler ShutdownHandler) {
	t.lock.Lock()
	defer t.lock.Unlock()
	if !t.alive {
		log.Printf("aborting handler registration due to lifecycle termination")
		handler(t.ctx)
		return
	}

	t.handlerList = append(t.handlerList, handler)
}

// Manually triggers a termination signal to end the application lifecycle.
//
// Uses SIGUSR1 (30) under the hood.
func (t *LifecycleToken) TerminateLifecycle() {
	t.signalChan <- syscall.SIGUSR1
}

// Retrieves the context associated with this [LifecycleToken]
func (t *LifecycleToken) GetContext() context.Context {
	return t.ctx
}

// Helper function meant to be run as a goroutine. This contains the lifecycle logic.
func (t *LifecycleToken) blockUntilTerminationSignal() {
	defer t.cancel()

	select {
	case sig := <-t.signalChan:
		log.Printf("received signal %d", sig)
	case <-t.ctx.Done():
		log.Printf("non-graceful shutdown detected")
	}

	t.lock.Lock()
	defer t.lock.Unlock()
	t.alive = false
	for _, handler := range t.handlerList {
		handler(t.ctx)
	}
	log.Printf("lifecycle complete")
}

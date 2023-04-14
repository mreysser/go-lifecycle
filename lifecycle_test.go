package lifecycle

import (
	"context"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLifecycleTermination(t *testing.T) {
	token = nil
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	testToken := GetLifecycleToken(ctx, nil)
	assert.NotNil(t, testToken)

	var stopped atomic.Bool
	var handled atomic.Bool

	go func() {
		for {
			if handled.Load() {
				break
			}
		}
		stopped.Store(true)
	}()

	// Normal sequence of events
	testToken.RegisterShutdownHandler(func(ctx context.Context) { handled.Store(true) })
	testToken.TerminateLifecycle()
	<-testToken.GetContext().Done()

	assert.True(t, stopped.Load())
	assert.True(t, handled.Load())
}

func TestLifecycleEarlyTermination(t *testing.T) {
	token = nil
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	testToken := GetLifecycleToken(ctx, []syscall.Signal{syscall.SIGTERM})
	assert.NotNil(t, testToken)

	var stopped atomic.Bool
	var handled atomic.Bool

	go func() {
		for {
			if handled.Load() {
				break
			}
		}
		stopped.Store(true)
	}()

	// Note out of order sequence of events
	testToken.TerminateLifecycle()
	<-token.GetContext().Done()
	testToken.RegisterShutdownHandler(func(ctx context.Context) { handled.Store(true) })

	// Artifical wait here, because testToken.GetContext().Done() is already closed.
	<-ctx.Done()

	assert.True(t, stopped.Load())
	assert.True(t, handled.Load())
}

func TestLifecycleContextCancelled(t *testing.T) {
	token = nil
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	testToken := GetLifecycleToken(ctx, []syscall.Signal{syscall.SIGTERM})
	assert.NotNil(t, testToken)

	var stopped atomic.Bool
	var handled atomic.Bool

	go func() {
		for {
			if handled.Load() {
				break
			}
		}
		stopped.Store(true)
	}()

	testToken.RegisterShutdownHandler(func(ctx context.Context) { handled.Store(true) })

	// Oops, context cancelled early
	cancel()

	// Artifical wait here, because testToken.GetContext().Done() is already closed.
	for {
		if stopped.Load() {
			break
		}
	}

	assert.True(t, stopped.Load())
	assert.True(t, handled.Load())
}

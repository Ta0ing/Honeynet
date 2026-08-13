//go:build !windows

package platformservice

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Run executes a component until SIGINT or SIGTERM is received.
func Run(_ string, component func(context.Context) error) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return component(ctx)
}

//go:build !windows

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func runServiceOrConsole() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runAPI(ctx)
}

// Command server is the HTTP API entry point. It boots the application
// container (see internal/app), starts the Fiber server and shuts down
// gracefully on SIGINT/SIGTERM.
package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang/internal/app"
)

func main() {
	container, shutdown := app.Boot()

	httpApp := app.NewHTTP(container)
	errCh := make(chan error, 1)

	go func() {
		if err := container.Serve(httpApp); err != nil && !errors.Is(err, context.Canceled) {
			errCh <- err
		} else {
			errCh <- nil
		}
	}()

	// Wait for an OS signal or a fatal server error.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			container.Logger.Error("server exited with error", "error", err.Error())
			_ = shutdown()
			os.Exit(1)
		}
	case sig := <-sigCh:
		container.Logger.Info("received shutdown signal", "signal", sig.String())
	}

	// Give in-flight requests a moment to finish before closing resources.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpApp.ShutdownWithContext(ctx)
	_ = shutdown()
}

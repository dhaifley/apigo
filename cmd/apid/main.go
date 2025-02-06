// apid is a service providing an API.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/dhaifley/apid"
)

// Main service entry point.
func main() {
	ctx := context.Background()

	svc := apid.New()

	if len(os.Args) > 0 && os.Args[1] == "version" {
		fmt.Println("apid", svc.Version())

		os.Exit(0)
	}

	if len(os.Args) > 0 && os.Args[1] == "migrate" {
		if err := svc.Migrate(ctx); err != nil {
			slog.Error("migrate error", "error", err)

			os.Exit(1)
		}

		os.Exit(0)
	}

	errCh := make(chan error, 1)

	go func(ctx context.Context, errCh chan error) {
		if err := svc.Start(ctx); err != nil {
			errCh <- err
		}
	}(ctx, errCh)

	ch := make(chan os.Signal, 1)

	signal.Notify(ch,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT, os.Interrupt)

	select {
	case <-ch:
		svc.Close(ctx)
	case err := <-errCh:
		slog.Error("server error", "error", err)

		os.Exit(1)
	}
}

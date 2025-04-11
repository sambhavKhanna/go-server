package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/sambhavKhanna/Infra/Logger"
)

func NewServer(
	logger logging.Logger,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, logger)

	var handler http.Handler = mux

	handler = LoggingMiddleware(handler, logger)
	return handler
}

func run(w io.Writer, ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	logger := logging.New(w)

	server := NewServer(logger)
	httpServer := &http.Server{
		Addr:         ":8080",
		Handler: 	server,
	}
	go func() {
		fmt.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10 * time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()

	wg.Wait()
	return nil
}

func main() {
	ctx := context.Background()
	if err := run(os.Stdout, ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

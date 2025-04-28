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

	"github.com/sambhavKhanna/infra/database"
	"github.com/sambhavKhanna/infra/logger"
	"github.com/sambhavKhanna/internal/service"
)

func run(w io.Writer, ctx context.Context, getenv func(string) string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	logger := logging.New(w)
	db, err := db.New(getenv)
	if err != nil {
		return err
	}

	server := service.NewServer(logger, db)
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: server,
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
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
		if err := db.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "error closing database: %s\n", err)
		}
	}()

	wg.Wait()
	return nil
}

func main() {
	getenv := func(key string) string {
		switch key {
		case "DB_HOST":
			return os.Getenv("POSTGRES_HOST")
		case "DB_PORT":
			return os.Getenv("POSTGRES_PORT")
		case "DB_USER":
			return os.Getenv("POSTGRES_USER")
		case "DB_PASSWORD":
			return os.Getenv("POSTGRES_PASSWORD")
		case "DB_NAME":
			return os.Getenv("POSTGRES_DB")
		default:
			return ""
		}
	}

	ctx := context.Background()
	if err := run(os.Stdout, ctx, getenv); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

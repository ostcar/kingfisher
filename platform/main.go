package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"webserver/database"
	"webserver/http"
	"webserver/roc"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := interruptContext()
	defer cancel()

	db := database.FileDB{File: "db.events"}

	reader, err := db.Reader()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	r, err := roc.New(reader)
	if err != nil {
		return fmt.Errorf("initial roc app: %w", err)
	}

	return http.Run(ctx, ":8090", r, db)
}

// interruptContext works like signal.NotifyContext
//
// In only listens on os.Interrupt. If the signal is received two times,
// os.Exit(1) is called.
func interruptContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		cancel()

		// If the signal was send for the second time, make a hard cut.
		<-sigint
		os.Exit(1)
	}()
	return ctx, cancel
}

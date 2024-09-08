package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"webserver/database"
	"webserver/http"
	"webserver/roc"

	"github.com/alecthomas/kong"
)

type cliConfig struct {
	Addr       string `help:"Address to listen on." default:":8090"`
	EventsFile string `help:"Path to the events file." default:"db.events" type:"path"`
}

func entry() {
	var cli cliConfig

	kong.Parse(&cli, kong.UsageOnError())

	if err := run(cli); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run(cli cliConfig) (err error) {
	ctx, cancel := interruptContext()
	defer cancel()

	db := database.FileDB{
		EventsFile: cli.EventsFile,
		EventType:  database.EventTypeLine, // TODO: create a config to set this
	}

	reader, err := db.EventsReader()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	r, err := roc.New(reader)
	if err != nil {
		return fmt.Errorf("initial roc app: %w", err)
	}

	return http.Run(ctx, cli.Addr, r, db)
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

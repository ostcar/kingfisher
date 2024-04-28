package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"webserver/database"
	"webserver/http"
	"webserver/roc"

	"github.com/alecthomas/kong"
)

func main() {
	var cli cliConfig

	kong.Parse(&cli, kong.UsageOnError())

	if err := run(cli); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

type cliConfig struct {
	Addr         string `help:"Address to listen on." default:":8090"`
	SnapshotFile string `help:"Path to the snapshot file." default:"db.snapshot" type:"path"`
	EventFile    string `help:"Path to the event file." default:"db.events" type:"path"`
}

func run(cli cliConfig) (err error) {
	ctx, cancel := interruptContext()
	defer cancel()

	db := database.FileDB{
		EventFile:    cli.EventFile,
		SnapshotFile: cli.SnapshotFile,
	}

	reader, err := db.EventReader()
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	initSnapshot, err := db.SnapshotRead()
	if err != nil {
		return fmt.Errorf("reading snapshot: %w", err)
	}

	r, err := roc.New(initSnapshot, reader)
	if err != nil {
		return fmt.Errorf("initial roc app: %w", err)
	}

	defer func() {
		if sErr := db.SnapshotWrite(r.DumpModel()); sErr != nil {
			err = errors.Join(err, fmt.Errorf("saving model snapshot: %w", sErr))
		}
	}()

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

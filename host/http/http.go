package http

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"webserver/database"
	"webserver/roc"
)

// Run starts the webserver
func Run(ctx context.Context, addr string, r *roc.Roc, db database.Database) error {
	srv := &http.Server{
		Addr:        addr,
		Handler:     handler(r, db),
		BaseContext: func(net.Listener) context.Context { return ctx },
	}

	// Shutdown logic in separate goroutine.
	wait := make(chan error)
	go func() {
		// Wait for the context to be closed.
		<-ctx.Done()

		if err := srv.Shutdown(context.Background()); err != nil {
			wait <- fmt.Errorf("HTTP server shutdown: %w", err)
			return
		}
		wait <- nil
	}()

	addrToLog := addr
	if strings.HasPrefix(addrToLog, ":") {
		addrToLog = "0.0.0.0" + addrToLog
	}
	log.Printf("Webserver is listening on: http://%s", addrToLog)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server failed: %v", err)
	}

	return <-wait
}

func handler(rocApp *roc.Roc, db database.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := rocApp.HanldeRequest(w, r, db.EventsWriter); err != nil {
			http.Error(w, "Error", 500)
			log.Printf("Error: %v", err)
			return
		}
	})
}

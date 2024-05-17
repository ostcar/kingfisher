package http

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

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

	log.Printf("Webserver is listening on: %s", addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server failed: %v", err)
	}

	return <-wait
}

func handler(rocApp *roc.Roc, db database.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var response roc.Response
		var err error

		request, err := roc.RequestFromHTTP(r)
		if err != nil {
			http.Error(w, "Error", 500)
			return
		}

		var outErr error
		if isWriteRequest(r.Method) {
			writer, err := db.RequestsWriter()
			if err != nil {
				http.Error(w, "Error", 500)
				return
			}

			response, outErr = rocApp.WriteRequest(request, writer)
		} else {
			response, outErr = rocApp.ReadRequest(request)
		}
		if outErr != nil {
			http.Error(w, "Error", 500)
			log.Printf("Error: %v", err)
			return
		}

		for _, header := range response.Headers {
			w.Header().Add(header.Name, header.Value)
		}
		w.WriteHeader(response.Status)

		w.Write([]byte(response.Body))
	})
}

func isWriteRequest(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

package database

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Database has the ability to read all data or append new.
type Database interface {
	EventsReader() (io.ReadCloser, error)
	EventsWriter() (io.WriteCloser, error)
}

// FileDB is a evet database based of one file.
type FileDB struct {
	EventsFile string
}

// EventsReader returns a reader to read the loged events from.
func (db FileDB) EventsReader() (io.ReadCloser, error) {
	f, err := os.Open(db.EventsFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("open events file: %w", err)
		}
		return io.NopCloser(strings.NewReader("")), nil

	}
	return f, nil
}

// EventsWriter returns a writer to store events.
func (db FileDB) EventsWriter() (io.WriteCloser, error) {
	f, err := os.OpenFile(db.EventsFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open events file: %w", err)
	}

	return f, nil
}

// MemoryDB stores the data in memory.
//
// Usefull for testing.
type MemoryDB struct {
	Requests string
	Snapshot string
}

// SnapshotRead returns the snapshot.
func (db *MemoryDB) SnapshotRead() ([]byte, error) {
	return []byte(db.Snapshot), nil
}

// SnapshotWrite stores the snapshot.
func (db *MemoryDB) SnapshotWrite(snapshot []byte) error {
	db.Snapshot = string(snapshot)
	return nil
}

// RequestsReader returns a reader to read the loged requests from.
func (db *MemoryDB) RequestsReader() (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(db.Requests)), nil
}

// RequestsWriter does currently nothing.
func (db *MemoryDB) RequestsWriter() (io.WriteCloser, error) {
	return NoopWriteCloser{}, nil
}

// NoopWriteCloser implements the io.WriteCloser interface, but does nothing.
type NoopWriteCloser struct{}

func (NoopWriteCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// Close is a noop
func (NoopWriteCloser) Close() error {
	return nil
}

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
	SnapshotRead() ([]byte, error)
	SnapshotWrite([]byte) error
	EventReader() (io.ReadCloser, error)
	EventWriter() (io.WriteCloser, error)
}

// FileDB is a evet database based of one file.
type FileDB struct {
	EventFile    string
	SnapshotFile string
}

// SnapshotRead reads the snapshot file.
//
// Returns nil, if the snapshot file does not exist.
func (db FileDB) SnapshotRead() ([]byte, error) {
	snaptop, err := os.ReadFile(db.SnapshotFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, fmt.Errorf("reading snapshot: %w", err)
	}
	return snaptop, nil
}

// SnapshotWrite writes the snapshot to the file.
//
// Also removes all Events.
func (db FileDB) SnapshotWrite(snapshot []byte) error {
	// TODO: Do not replace the file
	f, err := os.Create(db.SnapshotFile)
	if err != nil {
		return fmt.Errorf("create snapshot file: %w", err)
	}

	if _, err := f.Write(snapshot); err != nil {
		return fmt.Errorf("write snapshot: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("closing snaphot file: %w", err)
	}

	if err := os.Remove(db.EventFile); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("removing event file: %w", err)
		}
	}

	return nil
}

// EventReader returns a reader to read the events from.
func (db FileDB) EventReader() (io.ReadCloser, error) {
	f, err := os.Open(db.EventFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("open database file: %w", err)
		}
		return io.NopCloser(strings.NewReader("")), nil

	}
	return f, nil
}

// EventWriter returns a writer to store events
func (db FileDB) EventWriter() (io.WriteCloser, error) {
	f, err := os.OpenFile(db.EventFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open db file: %w", err)
	}

	return f, nil
}

// MemoryDB stores Events in memory.
//
// Usefull for testing.
type MemoryDB struct {
	Events   string
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

// EventReader returns a reader to read the events from.
func (db *MemoryDB) EventReader() (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(db.Events)), nil
}

// EventWriter does currently nothing.
func (db *MemoryDB) EventWriter() (io.WriteCloser, error) {
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

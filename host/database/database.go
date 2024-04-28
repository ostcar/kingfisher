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
	RequestsReader() (io.ReadCloser, error)
	RequestsWriter() (io.WriteCloser, error)
}

// FileDB is a evet database based of one file.
type FileDB struct {
	RequestsFile string
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
// Also clears the requests file.
func (db FileDB) SnapshotWrite(snapshot []byte) error {
	// TODO: Do not replace the file
	f, err := os.Create(db.SnapshotFile)
	if err != nil {
		return fmt.Errorf("creating snapshot file: %w", err)
	}

	if _, err := f.Write(snapshot); err != nil {
		return fmt.Errorf("writing snapshot: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("closing snaphot file: %w", err)
	}

	if err := os.Remove(db.RequestsFile); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("removing requests file: %w", err)
		}
	}

	return nil
}

// RequestsReader returns a reader to read the loged requests from.
func (db FileDB) RequestsReader() (io.ReadCloser, error) {
	f, err := os.Open(db.RequestsFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("open database file: %w", err)
		}
		return io.NopCloser(strings.NewReader("")), nil

	}
	return f, nil
}

// RequestsWriter returns a writer to store requests.
func (db FileDB) RequestsWriter() (io.WriteCloser, error) {
	f, err := os.OpenFile(db.RequestsFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open database file: %w", err)
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

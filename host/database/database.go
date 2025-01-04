package database

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Database has the ability to read all data or append new.
type Database interface {
	EventsReader() (func(yield func([]byte, error) bool), error)
	EventsWriter(event ...[]byte) error
}

// FileDB is a evet database based of one file.
type FileDB struct {
	EventsFile string
}

// EventsReader returns a reader to read the loged events from.
func (db FileDB) EventsReader() (func(yield func([]byte, error) bool), error) {
	f, err := os.Open(db.EventsFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("open events file: %w", err)
		}
		return func(yield func([]byte, error) bool) {}, nil
	}

	reader := bufio.NewReader(f)

	return func(yield func([]byte, error) bool) {
		defer f.Close()
		for {
			size, err := binary.ReadUvarint(reader)
			if err != nil {
				if errors.Is(err, io.EOF) || !yield(nil, err) {
					break
				}
			}

			buf := make([]byte, size)
			_, err = io.ReadFull(reader, buf)
			if !yield(buf, err) {
				break
			}
		}
	}, nil
}

// EventsWriter returns a writer to store events.
func (db FileDB) EventsWriter(eventList ...[]byte) error {
	f, err := os.OpenFile(db.EventsFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0o600)
	if err != nil {
		return fmt.Errorf("open events file: %w", err)
	}
	defer f.Close()

	for _, event := range eventList {
		buf := make([]byte, len(event)+8)
		written := binary.PutUvarint(buf, uint64(len(event)))
		copy(buf[written:], event)
		buf = buf[:written+len(event)]
		if _, err := f.Write(buf); err != nil {
			return fmt.Errorf("saving event: %w", err)
		}
	}
	return nil

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

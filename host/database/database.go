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

type EventType int

const (
	EventTypeLine = iota
	EventTypeBinary
	EventTypeText
)

// Database has the ability to read all data or append new.
type Database interface {
	EventsReader() (func(yield func([]byte, error) bool), error)
	EventsWriter(event ...[]byte) error
}

// FileDB is a evet database based of one file.
type FileDB struct {
	EventType  EventType
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

	switch db.EventType {
	case EventTypeLine:
		reader := bufio.NewReader(f)
		return func(yield func([]byte, error) bool) {
			for {
				line, err := reader.ReadBytes('\n')
				if errors.Is(err, io.EOF) {
					break
				}
				if !yield(line, err) {
					break
				}
			}
		}, nil

	default:
		return nil, fmt.Errorf("unknown eventtype %d", db.EventType)
	}

}

// EventsWriter returns a writer to store events.
func (db FileDB) EventsWriter(event ...[]byte) error {
	f, err := os.OpenFile(db.EventsFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0o600)
	if err != nil {
		return fmt.Errorf("open events file: %w", err)
	}
	defer f.Close()

	switch db.EventType {
	case EventTypeLine:
		// TODO: check, that event does not contain a newline
		for _, e := range event {
			if _, err := f.Write(e); err != nil {
				return fmt.Errorf("saving event: %w", err)
			}
			if _, err := f.Write([]byte("\n")); err != nil {
				return fmt.Errorf("saving newline: %w", err)
			}
		}
		return nil

	case EventTypeBinary:
		for _, e := range event {
			binary.Write(f, binary.LittleEndian, uint64(len(e)))
			if _, err := f.Write(e); err != nil {
				return fmt.Errorf("saving event: %w", err)
			}
		}
		return nil

	default:
		return fmt.Errorf("unknown eventtype %d", db.EventType)
	}
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

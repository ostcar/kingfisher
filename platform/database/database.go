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
	Reader() (io.ReadCloser, error)
	io.Writer
}

// FileDB is a evet database based of one file.
type FileDB struct {
	File string
}

// Reader opens the file and returns its reader.
func (db FileDB) Reader() (io.ReadCloser, error) {
	f, err := os.Open(db.File)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return io.NopCloser(strings.NewReader("")), nil
		}
		return nil, fmt.Errorf("open database file: %w", err)
	}
	return f, nil
}

// Write adds to the file
func (db FileDB) Write(p []byte) (int, error) {
	f, err := os.OpenFile(db.File, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return 0, fmt.Errorf("open db file: %w", err)
	}
	defer func() {
		wErr := f.Close()
		if err != nil {
			err = wErr
		}
	}()

	return f.Write(p)
}

// MemoryDB stores Events in memory.
//
// Usefull for testing.
type MemoryDB struct {
	Content string
}

// Reader reads the content.
func (db *MemoryDB) Reader() (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(db.Content)), nil
}

// Append adds a new event.
func (db *MemoryDB) Write(p []byte) (int, error) {
	db.Content += fmt.Sprintf("%s\n", p)
	return len(p), nil
}

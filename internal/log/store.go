package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var (
	enc = binary.BigEndian // defines the encoding for records
)

const (
	lenWidth = 8 // # of bytes used to store the record's length
)

type store struct {
	// Wrapper around a file with two APIs to append and read bytes
	*os.File
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64
}

func newStore(f *os.File) (*store, error) {
	// Create the store
	// check the file's size first (i.e. to continue using an existing store)
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	size := uint64(fi.Size())
	return &store{
		File: f,
		size: size,
		buf:  bufio.NewWriter(f),
	}, nil
}

func (s *store) Append(p []byte) (n uint64, pos uint64, err error) {
	// Append Method
	s.mu.Lock()
	defer s.mu.Unlock()
	pos = s.size // Knowing length of p makes it easier to read it later

	// Buffer the length of p to s.buf, to reduce number of system calls and improve performance
	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
		return 0, 0, err
	}

	// write to the file, register number of bytes written to w
	w, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}

	w += lenWidth
	s.size += uint64(w)
	return uint64(w), pos, nil
}

func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// flush the buffer, writing any buffered data to the file
	if err := s.buf.Flush(); err != nil {
		return nil, err
	}

	// the length of data is read and saved to size
	size := make([]byte, lenWidth)
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil {
		return nil, err
	}

	// fetch and return the record
	b := make([]byte, enc.Uint64(size))
	if _, err := s.File.ReadAt(b, int64(pos+lenWidth)); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return 0, err
	}
	return s.File.ReadAt(p, off)
}

func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.buf.Flush()
	if err != nil {
		return err
	}
	return s.File.Close()
}

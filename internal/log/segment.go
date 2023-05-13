package log

import (
	"fmt"
	"os"
	"path"

	api "github.com/magus-1/proglog/api/v1"
	"google.golang.org/protobuf/proto"
)

// Segment wraps the index and store types to coordinate operations
type segment struct {
	store                  *store
	index                  *index
	baseOffset, nextOffset uint64
	config                 Config
}

func newSegment(dir string, baseOffset uint64, c Config) (*segment, error) {
	// The log calls for a new segment (i.e. when the active segment hits max size)
	s := &segment{
		baseOffset: baseOffset,
		config:     c,
	}
	var err error

	// Open/Create the store file
	storeFile, err := os.OpenFile(
		path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".store")),
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, err
	}
	if s.store, err = newStore(storeFile); err != nil {
		return nil, err
	}

	// Open/Create the index file
	indexFile, err := os.OpenFile(
		path.Join(dir, fmt.Sprintf("%d%s", baseOffset, ".index")),
		os.O_RDWR|os.O_CREATE,
		0644,
	)
	if err != nil {
		return nil, err
	}
	if s.index, err = newIndex(indexFile, c); err != nil {
		return nil, err
	}
	if off, _, err := s.index.Read(-1); err != nil {
		// New index: the next record is the base offset
		s.nextOffset = baseOffset
	} else {
		// Existing index: the next record is current offset++
		s.nextOffset = baseOffset + uint64(off) + 1
	}
	return s, nil
}

func (s *segment) Append(record *api.Record) (offset uint64, err error) {
	// Writes the record to the segment, returns the offset (the log will return offset through API)
	cur := s.nextOffset
	record.Offset = cur
	p, err := proto.Marshal(record)
	if err != nil {
		return 0, err
	}

	// Append data to the store
	_, pos, err := s.store.Append(p)
	if err != nil {
		return 0, err
	}

	// Add an index entry
	if err = s.index.Write(
		// index offsets are relative to base offset
		uint32(s.nextOffset-uint64(s.baseOffset)),
		pos,
	); err != nil {
		return 0, err
	}
	s.nextOffset++
	return cur, nil
}

func (s *segment) Read(off uint64) (*api.Record, error) {
	// Return the record for the given offset
	// Get the relative offset from the given absolute index
	_, pos, err := s.index.Read(int64(off - s.baseOffset))
	if err != nil {
		return nil, err
	}

	// Read the record from the store
	p, err := s.store.Read(pos)
	if err != nil {
		return nil, err
	}

	// Return as protobuf
	record := &api.Record{}
	err = proto.Unmarshal(p, record)
	return record, err
}

func (s *segment) IsMaxed() bool {
	// Return true if either store or index are maxed out
	// Notice that either can be filled first, depending on Config and logs
	return s.store.size >= s.config.Segment.MaxStoreBytes ||
		s.index.size >= s.config.Segment.MaxIndexBytes
}

func (s *segment) Close() error {
	if err := s.index.Close(); err != nil {
		return err
	}
	if err := s.store.Close(); err != nil {
		return err
	}
	return nil
}

func (s *segment) Remove() error {
	if err := s.Close(); err != nil {
		return err
	}
	if err := os.Remove(s.index.Name()); err != nil {
		return err
	}
	if err := os.Remove(s.store.Name()); err != nil {
		return err
	}
	return nil
}

func nearestMultiple(j, k uint64) uint64 {
	// Tool to make sure we stay under the user's disk capacity
	if j >= 0 {
		return (j / k) * k
	}
	return ((j - k + 1) / k) * k

}

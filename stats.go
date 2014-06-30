package main

import (
	"fmt"
	"sync"
	"time"
)

// Stats holds current target's scarping statistics.
//
// New - number of new files encountered
// Old - number of existing files encountered
// Size - total size of downloaded data, in bytes
// Took - total time taken downloading
// Mtx - internal mutex
type Stats struct {
	New  int
	Old  int
	Size int64
	Took time.Duration
	Mtx  sync.Mutex
}

// addOld updates Stats for an existing file.
func (s *Stats) addOld() int {
	s.Mtx.Lock()
	defer s.Mtx.Unlock()

	s.Old++

	return s.Old
}

// addNew updates Stats for a new file.
func (s *Stats) addNew(n int64, took time.Duration) {
	s.Mtx.Lock()
	defer s.Mtx.Unlock()

	s.New++
	s.Size += n
	s.Took += took
}

// isEmpty check if there were any new files downloaded.
func (s *Stats) isEmpty() bool {
	s.Mtx.Lock()
	defer s.Mtx.Unlock()

	if s.New < 1 {
		return true
	}

	return false
}

// String formats Stats into a nice string.
func (s *Stats) String() string {
	s.Mtx.Lock()
	defer s.Mtx.Unlock()

	num := s.New
	size := float64(s.Size) / (1024.0 * 1024.0)
	speed := size / s.Took.Seconds()

	return fmt.Sprintf("%d files for %0.2f MB with avg. dl. speed %0.3f MB/s", num, size, speed)
}

// vim: ts=4 sw=4 sts=4

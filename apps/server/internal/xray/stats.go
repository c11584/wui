package xray

import (
	"sync/atomic"
)

// StatsTracker tracks traffic statistics
type StatsTracker struct {
	upload   atomic.Int64
	download atomic.Int64
}

func NewStatsTracker() *StatsTracker {
	return &StatsTracker{}
}

// AddUpload adds upload bytes
func (s *StatsTracker) AddUpload(bytes int64) {
	s.upload.Add(bytes)
}

// AddDownload adds download bytes
func (s *StatsTracker) AddDownload(bytes int64) {
	s.download.Add(bytes)
}

// GetStats returns current stats
func (s *StatsTracker) GetStats() (upload int64, download int64) {
	return s.upload.Load(), s.download.Load()
}

// Reset resets stats
func (s *StatsTracker) Reset() {
	s.upload.Store(0)
	s.download.Store(0)
}

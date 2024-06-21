// Package stats provides a thread-safe implementation for collecting and retrieving
// various statistics related to downloads, uploads, HTTP errors, and rate limiting.
//
// The Stats struct offers methods to increment counters for different events and
// retrieve aggregated statistics. It maintains both overall counts and statistics
// for the last 24 hours.
//
// Note: This package stores all data in memory. For high-load or long-running
// applications, consider persisting data to a database or using a more
// sophisticated time-series data structure.
package stats

import (
	"sync"
	"time"
)

type Stats struct {
	mu      sync.RWMutex
	data    StatsData
	last24h map[time.Time]StatsData
}

type StatsData struct {
	DownloadCount         int
	UploadCount           int
	Last24hDownloadCount  int
	Last24hUploadCount    int
	HTTPErrorCount        int
	Last24hHTTPErrorCount int
	RateLimitHitCount     int
	RateLimitedIPs        []RateLimitedIP
	StartTime             time.Time
}

type RateLimitedIP struct {
	IP           string
	RequestCount int
}

func NewStats() *Stats {
	return &Stats{
		last24h: make(map[time.Time]StatsData),
		data: StatsData{
			StartTime: time.Now(),
		},
	}
}

func (s *Stats) GetCurrentStats() StatsData {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.updateLast24hStats()
	return s.data
}

func (s *Stats) IncrementDownloads() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.DownloadCount++
	s.addToLast24h(func(sd *StatsData) {
		sd.DownloadCount++
	})
}

func (s *Stats) IncrementUploads() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.UploadCount++
	s.addToLast24h(func(sd *StatsData) {
		sd.UploadCount++
	})
}

func (s *Stats) IncrementHTTPErrors() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.HTTPErrorCount++
	s.addToLast24h(func(sd *StatsData) {
		sd.HTTPErrorCount++
	})
}

func (s *Stats) RecordRateLimitHit(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.RateLimitHitCount++
	found := false
	for i, rlIP := range s.data.RateLimitedIPs {
		if rlIP.IP == ip {
			s.data.RateLimitedIPs[i].RequestCount++
			found = true
			break
		}
	}
	if !found {
		s.data.RateLimitedIPs = append(s.data.RateLimitedIPs, RateLimitedIP{IP: ip, RequestCount: 1})
	}
}

func (s *Stats) addToLast24h(updateFunc func(*StatsData)) {
	now := time.Now().Truncate(time.Hour)
	if _, exists := s.last24h[now]; !exists {
		s.last24h[now] = StatsData{}
	}
	sd := s.last24h[now]
	updateFunc(&sd)
	s.last24h[now] = sd
}

func (s *Stats) updateLast24hStats() {
	now := time.Now()
	cutoff := now.Add(-24 * time.Hour)

	s.data.Last24hDownloadCount = 0
	s.data.Last24hUploadCount = 0
	s.data.Last24hHTTPErrorCount = 0

	for t, sd := range s.last24h {
		if t.Before(cutoff) {
			delete(s.last24h, t)
		} else {
			s.data.Last24hDownloadCount += sd.DownloadCount
			s.data.Last24hUploadCount += sd.UploadCount
			s.data.Last24hHTTPErrorCount += sd.HTTPErrorCount
		}
	}
}

func (s *Stats) GetUptime() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s == nil || s.data.StartTime.IsZero() {
		return 0
	}
	return time.Since(s.data.StartTime)
}

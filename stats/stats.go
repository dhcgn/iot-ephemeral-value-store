// Package stats provides a thread-safe implementation for collecting and retrieving
// various statistics related to downloads, uploads, HTTP errors, and rate limiting.
//
// The Stats struct offers methods to increment counters for different events and
// retrieve aggregated statistics. It maintains both overall counts and statistics
// for the last 24 hours.
//
// Note: This package stores all data in memory. For high-load or long-running
// applications, consider persisting data to a database or using a more
// sophisticated time-series data structure.package stats
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
	CountDownload int
	CountUpload   int

	CountLast24hDownload int
	CountLast24hUpload   int

	HTTPError        int
	HTTPLast24hError int

	CountRateLimitHits int
	RateLimitIPs       []RateLimitIP
}

type RateLimitIP struct {
	IP       string
	Requests int
}

func NewStats() *Stats {
	return &Stats{
		last24h: make(map[time.Time]StatsData),
	}
}

func (s *Stats) GetCurrentStats() StatsData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	s.updateLast24hStats()
	return s.data
}

func (s *Stats) IncrementDownload() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.CountDownload++
	s.addToLast24h(func(sd *StatsData) {
		sd.CountDownload++
	})
}

func (s *Stats) IncrementUpload() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.CountUpload++
	s.addToLast24h(func(sd *StatsData) {
		sd.CountUpload++
	})
}

func (s *Stats) IncrementHTTPError() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.HTTPError++
	s.addToLast24h(func(sd *StatsData) {
		sd.HTTPError++
	})
}

func (s *Stats) RecordRateLimitHit(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.CountRateLimitHits++
	found := false
	for i, rlIP := range s.data.RateLimitIPs {
		if rlIP.IP == ip {
			s.data.RateLimitIPs[i].Requests++
			found = true
			break
		}
	}
	if !found {
		s.data.RateLimitIPs = append(s.data.RateLimitIPs, RateLimitIP{IP: ip, Requests: 1})
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

	s.data.CountLast24hDownload = 0
	s.data.CountLast24hUpload = 0
	s.data.HTTPLast24hError = 0

	for t, sd := range s.last24h {
		if t.Before(cutoff) {
			delete(s.last24h, t)
		} else {
			s.data.CountLast24hDownload += sd.CountDownload
			s.data.CountLast24hUpload += sd.CountUpload
			s.data.HTTPLast24hError += sd.HTTPError
		}
	}
}

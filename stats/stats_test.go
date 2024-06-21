package stats

import (
	"testing"
	"time"
)

func TestNewStats(t *testing.T) {
	s := NewStats()
	if s == nil {
		t.Fatal("NewStats returned nil")
	}
	if s.data.StartTime.IsZero() {
		t.Error("StartTime was not set")
	}
	if len(s.last24h) != 0 {
		t.Error("last24h map was not initialized empty")
	}
}

func TestIncrementDownloads(t *testing.T) {
	s := NewStats()
	s.IncrementDownloads()
	if s.data.DownloadCount != 1 {
		t.Errorf("Expected DownloadCount to be 1, got %d", s.data.DownloadCount)
	}
}

func TestIncrementUploads(t *testing.T) {
	s := NewStats()
	s.IncrementUploads()
	if s.data.UploadCount != 1 {
		t.Errorf("Expected UploadCount to be 1, got %d", s.data.UploadCount)
	}
}

func TestIncrementHTTPErrors(t *testing.T) {
	s := NewStats()
	s.IncrementHTTPErrors()
	if s.data.HTTPErrorCount != 1 {
		t.Errorf("Expected HTTPErrorCount to be 1, got %d", s.data.HTTPErrorCount)
	}
}

func TestRecordRateLimitHit(t *testing.T) {
	s := NewStats()
	ip := "192.168.1.1"
	s.RecordRateLimitHit(ip)
	if s.data.RateLimitHitCount != 1 {
		t.Errorf("Expected RateLimitHitCount to be 1, got %d", s.data.RateLimitHitCount)
	}
	if len(s.data.RateLimitedIPs) != 1 {
		t.Errorf("Expected RateLimitedIPs to have 1 entry, got %d", len(s.data.RateLimitedIPs))
	}
	if s.data.RateLimitedIPs[0].IP != ip {
		t.Errorf("Expected RateLimitedIP to be %s, got %s", ip, s.data.RateLimitedIPs[0].IP)
	}
}

func TestGetCurrentStats(t *testing.T) {
	s := NewStats()
	s.IncrementDownloads()
	s.IncrementUploads()
	s.IncrementHTTPErrors()
	s.RecordRateLimitHit("192.168.1.1")

	stats := s.GetCurrentStats()
	if stats.DownloadCount != 1 {
		t.Errorf("Expected DownloadCount to be 1, got %d", stats.DownloadCount)
	}
	if stats.UploadCount != 1 {
		t.Errorf("Expected UploadCount to be 1, got %d", stats.UploadCount)
	}
	if stats.HTTPErrorCount != 1 {
		t.Errorf("Expected HTTPErrorCount to be 1, got %d", stats.HTTPErrorCount)
	}
	if stats.RateLimitHitCount != 1 {
		t.Errorf("Expected RateLimitHitCount to be 1, got %d", stats.RateLimitHitCount)
	}
}

func TestGetUptime(t *testing.T) {
	s := NewStats()
	time.Sleep(time.Second) // Wait for 1 second
	uptime := s.GetUptime()
	if uptime < time.Second {
		t.Errorf("Expected uptime to be at least 1 second, got %v", uptime)
	}
}

func TestLast24HourStats(t *testing.T) {
	s := NewStats()

	// Simulate events over the last 25 hours
	now := time.Now()
	for i := 0; i < 25; i++ {
		pastTime := now.Add(time.Duration(-i) * time.Hour)
		s.last24h[pastTime.Truncate(time.Hour)] = StatsData{
			DownloadCount:  1,
			UploadCount:    1,
			HTTPErrorCount: 1,
		}
	}

	stats := s.GetCurrentStats()
	if stats.Last24hDownloadCount != 24 {
		t.Errorf("Expected Last24hDownloadCount to be 24, got %d", stats.Last24hDownloadCount)
	}
	if stats.Last24hUploadCount != 24 {
		t.Errorf("Expected Last24hUploadCount to be 24, got %d", stats.Last24hUploadCount)
	}
	if stats.Last24hHTTPErrorCount != 24 {
		t.Errorf("Expected Last24hHTTPErrorCount to be 24, got %d", stats.Last24hHTTPErrorCount)
	}
}

func TestConcurrency(t *testing.T) {
	s := NewStats()
	concurrency := 100
	iterations := 1000

	done := make(chan bool)
	for i := 0; i < concurrency; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				s.IncrementDownloads()
				s.IncrementUploads()
				s.IncrementHTTPErrors()
				s.RecordRateLimitHit("192.168.1.1")
				s.GetCurrentStats()
				s.GetUptime()
			}
			done <- true
		}()
	}

	for i := 0; i < concurrency; i++ {
		<-done
	}

	stats := s.GetCurrentStats()
	expectedCount := concurrency * iterations
	if stats.DownloadCount != expectedCount {
		t.Errorf("Expected DownloadCount to be %d, got %d", expectedCount, stats.DownloadCount)
	}
	if stats.UploadCount != expectedCount {
		t.Errorf("Expected UploadCount to be %d, got %d", expectedCount, stats.UploadCount)
	}
	if stats.HTTPErrorCount != expectedCount {
		t.Errorf("Expected HTTPErrorCount to be %d, got %d", expectedCount, stats.HTTPErrorCount)
	}
	if stats.RateLimitHitCount != expectedCount {
		t.Errorf("Expected RateLimitHitCount to be %d, got %d", expectedCount, stats.RateLimitHitCount)
	}
}

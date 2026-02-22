package httphandler

import (
	"testing"
	"time"
)

func Test_addTimestampToThisData(t *testing.T) {
	t.Run("empty path adds per-key and global timestamps", func(t *testing.T) {
		// Truncate to second precision to match RFC3339 output
		before := time.Now().UTC().Truncate(time.Second)
		paramMap := map[string]string{
			"temp":     "23.5",
			"humidity": "45",
		}

		addTimestampToThisData(paramMap, "")

		after := time.Now().UTC().Add(time.Second)

		// Global timestamp must be present
		ts, ok := paramMap["timestamp"]
		if !ok {
			t.Fatal("expected 'timestamp' key to be set")
		}
		parsed, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			t.Fatalf("timestamp is not RFC3339: %v", err)
		}
		if parsed.Before(before) || parsed.After(after) {
			t.Errorf("timestamp %v is outside expected range [%v, %v]", parsed, before, after)
		}

		// Per-key timestamps must be present for the original keys
		for _, key := range []string{"temp", "humidity"} {
			tsKey := key + "_timestamp"
			tsVal, ok := paramMap[tsKey]
			if !ok {
				t.Errorf("expected per-key timestamp %q to be set", tsKey)
				continue
			}
			if _, err := time.Parse(time.RFC3339, tsVal); err != nil {
				t.Errorf("per-key timestamp %q is not RFC3339: %v", tsKey, err)
			}
		}
	})

	t.Run("non-empty path adds only global timestamp", func(t *testing.T) {
		before := time.Now().UTC().Truncate(time.Second)
		paramMap := map[string]string{
			"temp": "23.5",
		}

		addTimestampToThisData(paramMap, "room1")

		after := time.Now().UTC().Add(time.Second)

		// Global timestamp must be present
		ts, ok := paramMap["timestamp"]
		if !ok {
			t.Fatal("expected 'timestamp' key to be set")
		}
		parsed, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			t.Fatalf("timestamp is not RFC3339: %v", err)
		}
		if parsed.Before(before) || parsed.After(after) {
			t.Errorf("timestamp %v is outside expected range [%v, %v]", parsed, before, after)
		}

		// Per-key timestamp must NOT be set when path is non-empty
		if _, ok := paramMap["temp_timestamp"]; ok {
			t.Error("expected no per-key timestamp when path is non-empty")
		}
	})

	t.Run("empty paramMap with empty path still sets global timestamp", func(t *testing.T) {
		paramMap := map[string]string{}

		addTimestampToThisData(paramMap, "")

		if _, ok := paramMap["timestamp"]; !ok {
			t.Error("expected 'timestamp' key to be set even for empty paramMap")
		}
	})
}

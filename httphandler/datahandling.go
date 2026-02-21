package httphandler

import (
	"time"
)

func addTimestampToThisData(paramMap map[string]string, path string) {
	// if using patch and path is empty than add a timestamp with the value suffix
	if path == "" {
		allKeys := make([]string, 0, len(paramMap))
		for k := range paramMap {
			allKeys = append(allKeys, k)
		}
		// add a timestamp with the value suffix for all keys
		timestamp := time.Now().UTC().Format(time.RFC3339)
		for _, k := range allKeys {
			paramMap[k+"_timestamp"] = timestamp
		}
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	paramMap["timestamp"] = timestamp
}

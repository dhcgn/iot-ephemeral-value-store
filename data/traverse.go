package data

import (
	"fmt"
	"strings"
)

// TraverseField walks a nested map using the given slash-separated field path
// and returns the value at that location.
func TraverseField(data map[string]interface{}, fieldPath string) (interface{}, error) {
	keys := strings.Split(fieldPath, "/")
	var value interface{} = data

	for _, key := range keys {
		m, ok := value.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid parameter path: '%s'", fieldPath)
		}
		value, ok = m[key]
		if !ok {
			return nil, fmt.Errorf("parameter '%s' not found", fieldPath)
		}
	}

	return value, nil
}

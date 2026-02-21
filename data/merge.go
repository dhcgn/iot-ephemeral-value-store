package data

import "strings"

// MergeDataAtPath merges newData into existingData at the specified path.
// If path is empty, merges at root level.
// Path segments are separated by "/".
func MergeDataAtPath(existingData map[string]interface{}, path string, newData map[string]interface{}) {
	if path == "" {
		for k, v := range newData {
			existingData[k] = v
		}
		return
	}

	segments := strings.Split(path, "/")

	current := existingData
	for i, segment := range segments {
		if i == len(segments)-1 {
			existingMap, ok := current[segment].(map[string]interface{})
			if !ok {
				existingMap = make(map[string]interface{})
			}
			for k, v := range newData {
				existingMap[k] = v
			}
			current[segment] = existingMap
		} else {
			nextMap, ok := current[segment].(map[string]interface{})
			if !ok {
				nextMap = make(map[string]interface{})
				current[segment] = nextMap
			}
			current = nextMap
		}
	}
}

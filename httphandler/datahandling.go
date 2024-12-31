package httphandler

import (
	"strings"
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

func (c Config) modifyData(downloadKey string, paramMap map[string]string, path string, isPatch bool) (map[string]interface{}, error) {
	var dataToStore map[string]interface{}

	if isPatch {
		existingData, err := c.StorageInstance.Retrieve(downloadKey)
		if err != nil {
			return nil, err
		}
		mergeData(existingData, paramMap, strings.Split(path, "/"))
		dataToStore = existingData
	} else {
		dataToStore = make(map[string]interface{})
		for k, v := range paramMap {
			dataToStore[k] = v
		}
	}

	// add timestamp so that root level timestamp is always the latest of any updated value
	timestamp := time.Now().UTC().Format(time.RFC3339)
	dataToStore["timestamp"] = timestamp

	return dataToStore, nil
}

func mergeData(existingData map[string]interface{}, newData map[string]string, path []string) {
	if len(path) == 0 || (len(path) == 1 && path[0] == "") {
		for k, v := range newData {
			existingData[k] = v
		}
		return
	}

	currentKey := path[0]
	if _, exists := existingData[currentKey]; !exists {
		existingData[currentKey] = make(map[string]interface{})
	}

	if nestedMap, ok := existingData[currentKey].(map[string]interface{}); ok {
		mergeData(nestedMap, newData, path[1:])
	} else {
		existingData[currentKey] = newData
	}
}

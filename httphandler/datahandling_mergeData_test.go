package httphandler

import (
	"reflect"
	"testing"
)

func TestMergeData(t *testing.T) {
	tests := []struct {
		name         string
		existingData map[string]interface{}
		newData      map[string]string
		path         []string
		expectedData map[string]interface{}
	}{
		{
			name: "Empty path",
			existingData: map[string]interface{}{
				"key1": "value1",
			},
			newData: map[string]string{
				"key2": "value2",
			},
			path: []string{},
			expectedData: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "Single empty string in path",
			existingData: map[string]interface{}{
				"key1": "value1",
			},
			newData: map[string]string{
				"key2": "value2",
			},
			path: []string{""},
			expectedData: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "Nested merge",
			existingData: map[string]interface{}{
				"key1": map[string]interface{}{
					"subkey1": "subvalue1",
				},
			},
			newData: map[string]string{
				"subkey2": "subvalue2",
			},
			path: []string{"key1"},
			expectedData: map[string]interface{}{
				"key1": map[string]interface{}{
					"subkey1": "subvalue1",
					"subkey2": "subvalue2",
				},
			},
		},
		{
			name: "Override non-map value with new data",
			existingData: map[string]interface{}{
				"key1": "value1",
			},
			newData: map[string]string{
				"subkey1": "subvalue1",
			},
			path: []string{"key1"},
			expectedData: map[string]interface{}{
				"key1": map[string]string{
					"subkey1": "subvalue1",
				},
			},
		},
		{
			name: "Deeply nested merge",
			existingData: map[string]interface{}{
				"key1": map[string]interface{}{
					"subkey1": map[string]interface{}{
						"subsubkey1": "subsubvalue1",
					},
				},
			},
			newData: map[string]string{
				"subsubkey2": "subsubvalue2",
			},
			path: []string{"key1", "subkey1"},
			expectedData: map[string]interface{}{
				"key1": map[string]interface{}{
					"subkey1": map[string]interface{}{
						"subsubkey1": "subsubvalue1",
						"subsubkey2": "subsubvalue2",
					},
				},
			},
		},
		{
			name: "Create new nested map",
			existingData: map[string]interface{}{
				"key1": "value1",
			},
			newData: map[string]string{
				"subkey1": "subvalue1",
			},
			path: []string{"key2"},
			expectedData: map[string]interface{}{
				"key1": "value1",
				"key2": map[string]interface{}{
					"subkey1": "subvalue1",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mergeData(tt.existingData, tt.newData, tt.path)
			if !reflect.DeepEqual(tt.existingData, tt.expectedData) {
				t.Errorf("mergeData() = %v, want %v", tt.existingData, tt.expectedData)
			}
		})
	}
}

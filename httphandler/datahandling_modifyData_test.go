// datahandling_modifyData_test.go
package httphandler

import (
	"encoding/json"
	"testing"

	"github.com/dgraph-io/badger/v3"
	"github.com/stretchr/testify/assert"
)

type MockStorage struct {
	data map[string]map[string]interface{}
}

func (m *MockStorage) GetJSON(downloadKey string) ([]byte, error) {
	if data, exists := m.data[downloadKey]; exists {
		return json.Marshal(data)
	}
	return nil, badger.ErrKeyNotFound
}

func (m *MockStorage) Delete(downloadKey string) error {
	delete(m.data, downloadKey)
	return nil
}

func (m *MockStorage) Store(downloadKey string, dataToStore map[string]interface{}) error {
	m.data[downloadKey] = dataToStore
	return nil
}

func (m *MockStorage) Retrieve(downloadKey string) (map[string]interface{}, error) {
	if data, exists := m.data[downloadKey]; exists {
		return data, nil
	}
	return nil, badger.ErrKeyNotFound
}

func TestModifyData(t *testing.T) {
	mockStorage := &MockStorage{
		data: map[string]map[string]interface{}{
			"existingDownloadKey": {
				"field1": "value1",
			},
		},
	}

	config := Config{
		StorageInstance: mockStorage,
	}

	// Test case: Modify existing data with patch
	paramMap := map[string]string{
		"field2": "value2",
	}
	path := ""
	isPatch := true
	downloadKey := "existingDownloadKey"

	modifiedData, err := config.modifyData(downloadKey, paramMap, path, isPatch)
	assert.NoError(t, err)

	expectedData := map[string]interface{}{
		"field1":    "value1",
		"field2":    "value2",
		"timestamp": modifiedData["timestamp"],
	}
	assert.Equal(t, expectedData, modifiedData)

	// Test case: Modify existing data with patch and nested path
	paramMap = map[string]string{
		"nestedField": "nestedValue",
	}
	path = "nested"
	isPatch = true

	modifiedData, err = config.modifyData(downloadKey, paramMap, path, isPatch)
	assert.NoError(t, err)

	expectedData = map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
		"nested": map[string]interface{}{
			"nestedField": "nestedValue",
		},
		"timestamp": modifiedData["timestamp"],
	}
	assert.Equal(t, expectedData, modifiedData)

	// Test case: Overwrite data without patch
	paramMap = map[string]string{
		"newField": "newValue",
	}
	path = ""
	isPatch = false

	modifiedData, err = config.modifyData(downloadKey, paramMap, path, isPatch)
	assert.NoError(t, err)

	expectedData = map[string]interface{}{
		"newField":  "newValue",
		"timestamp": modifiedData["timestamp"],
	}
	assert.Equal(t, expectedData, modifiedData)
}

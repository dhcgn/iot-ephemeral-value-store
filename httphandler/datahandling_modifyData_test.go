package httphandler

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorage is a mock implementation of the Storage interface
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Retrieve(key string) (map[string]interface{}, error) {
	args := m.Called(key)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockStorage) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockStorage) GetJSON(key string) ([]byte, error) {
	args := m.Called(key)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockStorage) Store(key string, data map[string]interface{}) error {
	args := m.Called(key, data)
	return args.Error(0)
}

// Storage is an interface that defines the Retrieve method
type Storage interface {
	Retrieve(key string) (map[string]interface{}, error)
}

func TestModifyData(t *testing.T) {
	mockStorage := new(MockStorage)
	config := Config{StorageInstance: mockStorage}

	t.Run("isPatch true with existing data", func(t *testing.T) {
		downloadKey := "testKey"
		paramMap := map[string]string{"newKey": "newValue"}
		path := "some/path"
		isPatch := true

		existingData := map[string]interface{}{"existingKey": "existingValue"}
		mockStorage.On("Retrieve", downloadKey).Return(existingData, nil)

		result, err := config.modifyData(downloadKey, paramMap, path, isPatch)
		assert.NoError(t, err)
		assert.Equal(t, "existingValue", result["existingKey"])
		assert.Equal(t, "newValue", result["newKey"])
		assert.NotEmpty(t, result["timestamp"])
	})

	t.Run("isPatch false creates new data", func(t *testing.T) {
		downloadKey := "testKey"
		paramMap := map[string]string{"newKey": "newValue"}
		path := "some/path"
		isPatch := false

		result, err := config.modifyData(downloadKey, paramMap, path, isPatch)
		assert.NoError(t, err)
		assert.Equal(t, "newValue", result["newKey"])
		assert.NotEmpty(t, result["timestamp"])
	})

	t.Run("isPatch true with error retrieving data", func(t *testing.T) {
		downloadKey := "testKey"
		paramMap := map[string]string{"newKey": "newValue"}
		path := "some/path"
		isPatch := true

		mockStorage.On("Retrieve", downloadKey).Return(nil, errors.New("retrieve error"))

		result, err := config.modifyData(downloadKey, paramMap, path, isPatch)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

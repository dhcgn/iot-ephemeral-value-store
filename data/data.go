package data

import (
	"fmt"
	"time"

	"github.com/dhcgn/iot-ephemeral-value-store/domain"
	"github.com/dhcgn/iot-ephemeral-value-store/storage"
)

// Service encapsulates the shared business logic for data operations.
// Both httphandler and mcphandler delegate to this service.
type Service struct {
	StorageInstance storage.Storage
}

// GenerateKeyPair generates a new upload/download key pair.
func (s *Service) GenerateKeyPair() (uploadKey, downloadKey string, err error) {
	uploadKey = domain.GenerateRandomKey()
	downloadKey, err = domain.DeriveDownloadKey(uploadKey)
	if err != nil {
		return "", "", fmt.Errorf("error deriving download key: %w", err)
	}
	return uploadKey, downloadKey, nil
}

// Upload validates the upload key, replaces all data with the given params
// (adding a root timestamp), and stores it. Returns the download key and stored data.
func (s *Service) Upload(uploadKey string, params map[string]string) (downloadKey string, storedData map[string]interface{}, err error) {
	if err := domain.ValidateUploadKey(uploadKey); err != nil {
		return "", nil, fmt.Errorf("invalid upload key: %w", err)
	}

	downloadKey, err = domain.DeriveDownloadKey(uploadKey)
	if err != nil {
		return "", nil, fmt.Errorf("error deriving download key: %w", err)
	}

	data := make(map[string]interface{})
	for k, v := range params {
		data[k] = v
	}
	data["timestamp"] = time.Now().UTC().Format(time.RFC3339)

	if err := s.StorageInstance.Store(downloadKey, data); err != nil {
		return "", nil, fmt.Errorf("error storing data: %w", err)
	}

	return downloadKey, data, nil
}

// Patch validates the upload key, retrieves existing data, merges the given
// params at the specified path, adds a root timestamp, and stores the result.
func (s *Service) Patch(uploadKey string, path string, params map[string]string) (downloadKey string, storedData map[string]interface{}, err error) {
	if err := domain.ValidateUploadKey(uploadKey); err != nil {
		return "", nil, fmt.Errorf("invalid upload key: %w", err)
	}

	downloadKey, err = domain.DeriveDownloadKey(uploadKey)
	if err != nil {
		return "", nil, fmt.Errorf("error deriving download key: %w", err)
	}

	existingData, err := s.StorageInstance.Retrieve(downloadKey)
	if err != nil {
		return "", nil, fmt.Errorf("error retrieving existing data: %w", err)
	}

	newData := make(map[string]interface{})
	for k, v := range params {
		newData[k] = v
	}

	MergeDataAtPath(existingData, path, newData)

	existingData["timestamp"] = time.Now().UTC().Format(time.RFC3339)

	if err := s.StorageInstance.Store(downloadKey, existingData); err != nil {
		return "", nil, fmt.Errorf("error storing data: %w", err)
	}

	return downloadKey, existingData, nil
}

// DownloadJSON retrieves the raw JSON bytes for the given download key.
func (s *Service) DownloadJSON(downloadKey string) ([]byte, error) {
	jsonData, err := s.StorageInstance.GetJSON(downloadKey)
	if err != nil {
		return nil, fmt.Errorf("invalid download key or data not found: %w", err)
	}
	return jsonData, nil
}

// DownloadField retrieves a specific field from the stored data by traversing
// the nested map using the given slash-separated field path.
func (s *Service) DownloadField(downloadKey string, fieldPath string) (interface{}, error) {
	data, err := s.StorageInstance.Retrieve(downloadKey)
	if err != nil {
		return nil, fmt.Errorf("invalid download key or data not found: %w", err)
	}

	value, err := TraverseField(data, fieldPath)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// Delete validates the upload key and deletes the associated data.
func (s *Service) Delete(uploadKey string) (downloadKey string, err error) {
	if err := domain.ValidateUploadKey(uploadKey); err != nil {
		return "", fmt.Errorf("invalid upload key: %w", err)
	}

	downloadKey, err = domain.DeriveDownloadKey(uploadKey)
	if err != nil {
		return "", fmt.Errorf("error deriving download key: %w", err)
	}

	if err := s.StorageInstance.Delete(downloadKey); err != nil {
		return "", fmt.Errorf("error deleting data: %w", err)
	}

	return downloadKey, nil
}

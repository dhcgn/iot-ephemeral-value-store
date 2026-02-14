package mcphandler

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/dhcgn/iot-ephemeral-value-store/domain"
	"github.com/dhcgn/iot-ephemeral-value-store/stats"
	"github.com/dhcgn/iot-ephemeral-value-store/storage"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestGenerateKeyPairHandler(t *testing.T) {
	// Setup
	storageInstance := storage.NewInMemoryStorage()
	statsInstance := stats.NewStats()
	config := Config{
		StorageInstance: storageInstance,
		StatsInstance:   statsInstance,
		ServerHost:      "http://localhost:8080",
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := &GenerateKeyPairInput{}

	// Execute
	result, _, err := config.GenerateKeyPairHandler(ctx, req, input)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	if len(result.Content) == 0 {
		t.Fatal("Expected content in result")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected text content")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &data); err != nil {
		t.Fatalf("Expected valid JSON, got error: %v", err)
	}

	if data["upload_key"] == nil {
		t.Error("Expected upload_key in response")
	}

	if data["download_key"] == nil {
		t.Error("Expected download_key in response")
	}
}

func TestUploadDataHandler(t *testing.T) {
	// Setup
	storageInstance := storage.NewInMemoryStorage()
	statsInstance := stats.NewStats()
	config := Config{
		StorageInstance: storageInstance,
		StatsInstance:   statsInstance,
		ServerHost:      "http://localhost:8080",
	}

	uploadKey := domain.GenerateRandomKey()
	downloadKey, _ := domain.DeriveDownloadKey(uploadKey)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := &UploadDataInput{
		UploadKey: uploadKey,
		Parameters: map[string]string{
			"temp":     "23.5",
			"humidity": "45",
		},
	}

	// Execute
	result, _, err := config.UploadDataHandler(ctx, req, input)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	// Verify data was stored
	storedData, err := storageInstance.GetJSON(downloadKey)
	if err != nil {
		t.Fatalf("Expected data to be stored, got error: %v", err)
	}

	var dataMap map[string]interface{}
	json.Unmarshal(storedData, &dataMap)

	if dataMap["temp"] != "23.5" {
		t.Errorf("Expected temp=23.5, got %v", dataMap["temp"])
	}

	if dataMap["humidity"] != "45" {
		t.Errorf("Expected humidity=45, got %v", dataMap["humidity"])
	}

	if dataMap["timestamp"] == nil {
		t.Error("Expected timestamp to be set")
	}
}

func TestPatchDataHandler(t *testing.T) {
	// Setup
	storageInstance := storage.NewInMemoryStorage()
	statsInstance := stats.NewStats()
	config := Config{
		StorageInstance: storageInstance,
		StatsInstance:   statsInstance,
		ServerHost:      "http://localhost:8080",
	}

	uploadKey := domain.GenerateRandomKey()
	downloadKey, _ := domain.DeriveDownloadKey(uploadKey)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// First upload
	input1 := &PatchDataInput{
		UploadKey: uploadKey,
		Path:      "room1",
		Parameters: map[string]string{
			"temp": "20",
		},
	}
	config.PatchDataHandler(ctx, req, input1)

	// Second upload (patch)
	input2 := &PatchDataInput{
		UploadKey: uploadKey,
		Path:      "room2",
		Parameters: map[string]string{
			"temp": "22",
		},
	}
	result, _, err := config.PatchDataHandler(ctx, req, input2)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	// Verify data was merged
	storedData, err := storageInstance.GetJSON(downloadKey)
	if err != nil {
		t.Fatalf("Expected data to be stored, got error: %v", err)
	}

	var dataMap map[string]interface{}
	json.Unmarshal(storedData, &dataMap)

	// Check nested structure
	room1, ok := dataMap["room1"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected room1 to be a map")
	}
	if room1["temp"] != "20" {
		t.Errorf("Expected room1 temp=20, got %v", room1["temp"])
	}

	room2, ok := dataMap["room2"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected room2 to be a map")
	}
	if room2["temp"] != "22" {
		t.Errorf("Expected room2 temp=22, got %v", room2["temp"])
	}
}

func TestDownloadDataHandler(t *testing.T) {
	// Setup
	storageInstance := storage.NewInMemoryStorage()
	statsInstance := stats.NewStats()
	config := Config{
		StorageInstance: storageInstance,
		StatsInstance:   statsInstance,
		ServerHost:      "http://localhost:8080",
	}

	uploadKey := domain.GenerateRandomKey()
	downloadKey, _ := domain.DeriveDownloadKey(uploadKey)

	// Upload some data first
	data := map[string]interface{}{
		"temp":     "23.5",
		"humidity": "45",
		"room": map[string]interface{}{
			"temp": "22",
		},
	}
	storageInstance.Store(downloadKey, data)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	t.Run("Download all data", func(t *testing.T) {
		input := &DownloadDataInput{
			DownloadKey: downloadKey,
		}

		result, _, err := config.DownloadDataHandler(ctx, req, input)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result == nil {
			t.Fatal("Expected result to be non-nil")
		}

		textContent, ok := result.Content[0].(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected text content")
		}

		var responseData map[string]interface{}
		if err := json.Unmarshal([]byte(textContent.Text), &responseData); err != nil {
			t.Fatalf("Expected valid JSON, got error: %v", err)
		}

		if responseData["data"] == nil {
			t.Error("Expected data field in response")
		}
	})

	t.Run("Download specific parameter", func(t *testing.T) {
		input := &DownloadDataInput{
			DownloadKey: downloadKey,
			Parameter:   "temp",
		}

		result, _, err := config.DownloadDataHandler(ctx, req, input)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result == nil {
			t.Fatal("Expected result to be non-nil")
		}
	})

	t.Run("Download nested parameter", func(t *testing.T) {
		input := &DownloadDataInput{
			DownloadKey: downloadKey,
			Parameter:   "room/temp",
		}

		result, _, err := config.DownloadDataHandler(ctx, req, input)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if result == nil {
			t.Fatal("Expected result to be non-nil")
		}
	})

	t.Run("Download non-existent parameter", func(t *testing.T) {
		input := &DownloadDataInput{
			DownloadKey: downloadKey,
			Parameter:   "nonexistent",
		}

		_, _, err := config.DownloadDataHandler(ctx, req, input)

		if err == nil {
			t.Error("Expected error for non-existent parameter")
		}
	})
}

func TestDeleteDataHandler(t *testing.T) {
	// Setup
	storageInstance := storage.NewInMemoryStorage()
	statsInstance := stats.NewStats()
	config := Config{
		StorageInstance: storageInstance,
		StatsInstance:   statsInstance,
		ServerHost:      "http://localhost:8080",
	}

	uploadKey := domain.GenerateRandomKey()
	downloadKey, _ := domain.DeriveDownloadKey(uploadKey)

	// Upload some data first
	data := map[string]interface{}{
		"temp": "23.5",
	}
	storageInstance.Store(downloadKey, data)

	ctx := context.Background()
	req := &mcp.CallToolRequest{}
	input := &DeleteDataInput{
		UploadKey: uploadKey,
	}

	// Execute
	result, _, err := config.DeleteDataHandler(ctx, req, input)

	// Verify
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	// Verify data was deleted
	_, err = storageInstance.GetJSON(downloadKey)
	if err == nil {
		t.Error("Expected data to be deleted")
	}
}

func TestInvalidUploadKey(t *testing.T) {
	// Setup
	storageInstance := storage.NewInMemoryStorage()
	statsInstance := stats.NewStats()
	config := Config{
		StorageInstance: storageInstance,
		StatsInstance:   statsInstance,
		ServerHost:      "http://localhost:8080",
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	t.Run("Invalid upload key in upload", func(t *testing.T) {
		input := &UploadDataInput{
			UploadKey: "invalid",
			Parameters: map[string]string{
				"temp": "23.5",
			},
		}

		_, _, err := config.UploadDataHandler(ctx, req, input)

		if err == nil {
			t.Error("Expected error for invalid upload key")
		}
	})

	t.Run("Invalid upload key in patch", func(t *testing.T) {
		input := &PatchDataInput{
			UploadKey: "invalid",
			Parameters: map[string]string{
				"temp": "23.5",
			},
		}

		_, _, err := config.PatchDataHandler(ctx, req, input)

		if err == nil {
			t.Error("Expected error for invalid upload key")
		}
	})

	t.Run("Invalid upload key in delete", func(t *testing.T) {
		input := &DeleteDataInput{
			UploadKey: "invalid",
		}

		_, _, err := config.DeleteDataHandler(ctx, req, input)

		if err == nil {
			t.Error("Expected error for invalid upload key")
		}
	})
}

func TestMergeDataAtPath(t *testing.T) {
	t.Run("Merge at root", func(t *testing.T) {
		existing := map[string]interface{}{
			"a": "1",
		}
		new := map[string]interface{}{
			"b": "2",
		}

		result := mergeDataAtPath(existing, "", new)

		if result["a"] != "1" {
			t.Errorf("Expected a=1, got %v", result["a"])
		}
		if result["b"] != "2" {
			t.Errorf("Expected b=2, got %v", result["b"])
		}
	})

	t.Run("Merge at nested path", func(t *testing.T) {
		existing := map[string]interface{}{
			"room1": map[string]interface{}{
				"temp": "20",
			},
		}
		new := map[string]interface{}{
			"humidity": "45",
		}

		result := mergeDataAtPath(existing, "room1", new)

		room1, ok := result["room1"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected room1 to be a map")
		}
		if room1["temp"] != "20" {
			t.Errorf("Expected temp=20, got %v", room1["temp"])
		}
		if room1["humidity"] != "45" {
			t.Errorf("Expected humidity=45, got %v", room1["humidity"])
		}
	})

	t.Run("Create nested path", func(t *testing.T) {
		existing := map[string]interface{}{}
		new := map[string]interface{}{
			"temp": "22",
		}

		result := mergeDataAtPath(existing, "room1/sensors", new)

		room1, ok := result["room1"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected room1 to be a map")
		}
		sensors, ok := room1["sensors"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected sensors to be a map")
		}
		if sensors["temp"] != "22" {
			t.Errorf("Expected temp=22, got %v", sensors["temp"])
		}
	})
}

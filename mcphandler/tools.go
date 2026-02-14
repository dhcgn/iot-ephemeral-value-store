package mcphandler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/dhcgn/iot-ephemeral-value-store/domain"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GenerateKeyPairInput represents the input for generating a key pair
type GenerateKeyPairInput struct{}

// UploadDataInput represents the input for uploading data
type UploadDataInput struct {
	UploadKey  string            `json:"upload_key" jsonschema:"The upload key (256-bit hex string)"`
	Parameters map[string]string `json:"parameters" jsonschema:"Key-value pairs to upload"`
}

// PatchDataInput represents the input for patching data
type PatchDataInput struct {
	UploadKey  string            `json:"upload_key" jsonschema:"The upload key (256-bit hex string)"`
	Path       string            `json:"path" jsonschema:"Nested path for the data (e.g. 'room1/sensors' creates nested structure)"`
	Parameters map[string]string `json:"parameters" jsonschema:"Key-value pairs to merge at the specified path"`
}

// DownloadDataInput represents the input for downloading data
type DownloadDataInput struct {
	DownloadKey string `json:"download_key" jsonschema:"The download key to retrieve data"`
	Parameter   string `json:"parameter,omitempty" jsonschema:"Optional parameter path to retrieve (e.g. 'temp' or 'room1/temp'). If not provided returns all data as JSON"`
}

// DeleteDataInput represents the input for deleting data
type DeleteDataInput struct {
	UploadKey string `json:"upload_key" jsonschema:"The upload key for the data to delete"`
}

// GenerateKeyPairHandler handles the generation of upload/download key pairs
func (c Config) GenerateKeyPairHandler(ctx context.Context, req *mcp.CallToolRequest, params *GenerateKeyPairInput) (*mcp.CallToolResult, any, error) {
	// Check context cancellation
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}

	uploadKey := domain.GenerateRandomKey()
	downloadKey, err := domain.DeriveDownloadKey(uploadKey)
	if err != nil {
		c.StatsInstance.IncrementHTTPErrors()
		return nil, nil, fmt.Errorf("error deriving download key: %w", err)
	}

	result := map[string]interface{}{
		"upload_key":   uploadKey,
		"download_key": downloadKey,
		"upload_url":   fmt.Sprintf("%s/u/%s?param=value", c.ServerHost, uploadKey),
		"download_url": fmt.Sprintf("%s/d/%s/json", c.ServerHost, downloadKey),
		"message":      "Key pair generated successfully. Use the upload key to store data and the download key to retrieve it. The upload key must be kept secret.",
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, nil, nil
}

// UploadDataHandler handles data upload
func (c Config) UploadDataHandler(ctx context.Context, req *mcp.CallToolRequest, params *UploadDataInput) (*mcp.CallToolResult, any, error) {
	// Check context cancellation
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}

	// Validate upload key
	if err := domain.ValidateUploadKey(params.UploadKey); err != nil {
		return nil, nil, fmt.Errorf("invalid upload key: %w", err)
	}

	// Derive download key
	downloadKey, err := domain.DeriveDownloadKey(params.UploadKey)
	if err != nil {
		c.StatsInstance.IncrementHTTPErrors()
		return nil, nil, fmt.Errorf("error deriving download key: %w", err)
	}

	// Add timestamp
	data := make(map[string]interface{})
	for k, v := range params.Parameters {
		data[k] = v
	}
	data["timestamp"] = time.Now().UTC().Format(time.RFC3339)

	// Store data (replaces existing data)
	if err := c.StorageInstance.Store(downloadKey, data); err != nil {
		return nil, nil, fmt.Errorf("error storing data: %w", err)
	}
	c.StatsInstance.IncrementUploads()

	// Build parameter URLs
	paramURLs := make(map[string]string)
	for k := range params.Parameters {
		paramURLs[k] = fmt.Sprintf("%s/d/%s/plain/%s", c.ServerHost, downloadKey, url.PathEscape(k))
	}

	result := map[string]interface{}{
		"message":         "Data uploaded successfully",
		"download_url":    fmt.Sprintf("%s/d/%s/json", c.ServerHost, downloadKey),
		"parameter_urls":  paramURLs,
		"parameter_count": len(params.Parameters),
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, nil, nil
}

// PatchDataHandler handles merging data into existing storage
func (c Config) PatchDataHandler(ctx context.Context, req *mcp.CallToolRequest, params *PatchDataInput) (*mcp.CallToolResult, any, error) {
	// Check context cancellation
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}

	// Validate upload key
	if err := domain.ValidateUploadKey(params.UploadKey); err != nil {
		return nil, nil, fmt.Errorf("invalid upload key: %w", err)
	}

	// Derive download key
	downloadKey, err := domain.DeriveDownloadKey(params.UploadKey)
	if err != nil {
		c.StatsInstance.IncrementHTTPErrors()
		return nil, nil, fmt.Errorf("error deriving download key: %w", err)
	}

	// Get existing data
	var existingData map[string]interface{}
	existingJSON, err := c.StorageInstance.GetJSON(downloadKey)
	if err == nil {
		// Data exists, unmarshal it
		if err := json.Unmarshal(existingJSON, &existingData); err != nil {
			return nil, nil, fmt.Errorf("error unmarshaling existing data: %w", err)
		}
	} else {
		// No existing data, start fresh
		existingData = make(map[string]interface{})
	}

	// Prepare new data with timestamp
	newData := make(map[string]interface{})
	for k, v := range params.Parameters {
		newData[k] = v
	}
	newData["timestamp"] = time.Now().UTC().Format(time.RFC3339)

	// Merge data at the specified path
	mergedData := mergeDataAtPath(existingData, params.Path, newData)

	// Update the root timestamp
	mergedData["timestamp"] = time.Now().UTC().Format(time.RFC3339)

	// Store merged data
	if err := c.StorageInstance.Store(downloadKey, mergedData); err != nil {
		return nil, nil, fmt.Errorf("error storing data: %w", err)
	}
	c.StatsInstance.IncrementUploads()

	// Build parameter URLs
	paramURLs := make(map[string]string)
	var pathPrefix string
	if params.Path != "" {
		pathPrefix = params.Path + "/"
	}
	for k := range params.Parameters {
		paramPath := pathPrefix + k
		paramURLs[k] = fmt.Sprintf("%s/d/%s/plain/%s", c.ServerHost, downloadKey, url.PathEscape(paramPath))
	}

	result := map[string]interface{}{
		"message":         "Data merged successfully",
		"download_url":    fmt.Sprintf("%s/d/%s/json", c.ServerHost, downloadKey),
		"parameter_urls":  paramURLs,
		"path":            params.Path,
		"parameter_count": len(params.Parameters),
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, nil, nil
}

// DownloadDataHandler handles data retrieval
func (c Config) DownloadDataHandler(ctx context.Context, req *mcp.CallToolRequest, params *DownloadDataInput) (*mcp.CallToolResult, any, error) {
	// Check context cancellation
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}

	// Get data from storage
	jsonData, err := c.StorageInstance.GetJSON(params.DownloadKey)
	if err != nil {
		c.StatsInstance.IncrementHTTPErrors()
		return nil, nil, fmt.Errorf("invalid download key or data not found: %w", err)
	}

	// Parse JSON data
	var dataMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &dataMap); err != nil {
		c.StatsInstance.IncrementHTTPErrors()
		return nil, nil, fmt.Errorf("error decoding JSON: %w", err)
	}

	c.StatsInstance.IncrementDownloads()

	var result map[string]interface{}

	// If no parameter specified, return all data
	if params.Parameter == "" {
		result = map[string]interface{}{
			"data":    dataMap,
			"message": "Retrieved all data as JSON",
		}
	} else {
		// Extract specific parameter (supports nested paths like "room1/temp")
		keys := strings.Split(params.Parameter, "/")
		var value interface{} = dataMap

		for _, key := range keys {
			if m, ok := value.(map[string]interface{}); ok {
				value, ok = m[key]
				if !ok {
					c.StatsInstance.IncrementHTTPErrors()
					return nil, nil, fmt.Errorf("parameter '%s' not found", params.Parameter)
				}
			} else {
				c.StatsInstance.IncrementHTTPErrors()
				return nil, nil, fmt.Errorf("invalid parameter path: '%s'", params.Parameter)
			}
		}

		result = map[string]interface{}{
			"data":      value,
			"parameter": params.Parameter,
			"message":   fmt.Sprintf("Retrieved parameter '%s'", params.Parameter),
		}
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, nil, nil
}

// DeleteDataHandler handles data deletion
func (c Config) DeleteDataHandler(ctx context.Context, req *mcp.CallToolRequest, params *DeleteDataInput) (*mcp.CallToolResult, any, error) {
	// Check context cancellation
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}

	// Validate upload key
	if err := domain.ValidateUploadKey(params.UploadKey); err != nil {
		return nil, nil, fmt.Errorf("invalid upload key: %w", err)
	}

	// Derive download key
	downloadKey, err := domain.DeriveDownloadKey(params.UploadKey)
	if err != nil {
		c.StatsInstance.IncrementHTTPErrors()
		return nil, nil, fmt.Errorf("error deriving download key: %w", err)
	}

	// Delete the data
	if err := c.StorageInstance.Delete(downloadKey); err != nil {
		return nil, nil, fmt.Errorf("error deleting data: %w", err)
	}

	result := map[string]interface{}{
		"message": "Data deleted successfully",
		"success": true,
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, nil, nil
}

// mergeDataAtPath merges newData into existingData at the specified path
// If path is empty, merges at root level
// Path segments are separated by "/"
func mergeDataAtPath(existingData map[string]interface{}, path string, newData map[string]interface{}) map[string]interface{} {
	if path == "" {
		// Merge at root level
		for k, v := range newData {
			existingData[k] = v
		}
		return existingData
	}

	// Split path into segments
	segments := strings.Split(path, "/")

	// Navigate/create nested structure
	current := existingData
	for i, segment := range segments {
		if i == len(segments)-1 {
			// Last segment - merge the data here
			existingMap, ok := current[segment].(map[string]interface{})
			if !ok {
				existingMap = make(map[string]interface{})
			}
			for k, v := range newData {
				existingMap[k] = v
			}
			current[segment] = existingMap
		} else {
			// Intermediate segment - ensure it's a map
			nextMap, ok := current[segment].(map[string]interface{})
			if !ok {
				nextMap = make(map[string]interface{})
				current[segment] = nextMap
			}
			current = nextMap
		}
	}

	return existingData
}

// RegisterTools registers all MCP tools with the server
func (c Config) RegisterTools(server *mcp.Server) {
	// Tool: generate_key_pair
	mcp.AddTool(server, &mcp.Tool{
		Name:        "generate_key_pair",
		Description: "Generate a new upload/download key pair for the IoT ephemeral value store. The upload key is used to store data (must be kept secret), and the download key is used to retrieve data (can be shared for read-only access). Keys are cryptographically linked - the download key is derived from the upload key using SHA256.",
	}, c.GenerateKeyPairHandler)

	// Tool: upload_data
	mcp.AddTool(server, &mcp.Tool{
		Name:        "upload_data",
		Description: "Upload data to the IoT ephemeral value store using an upload key. This operation REPLACES any existing data for this key with the new data. All parameters are stored with an automatic timestamp. Use this for initial data uploads or complete replacements. For merging with existing data, use patch_data instead.",
	}, c.UploadDataHandler)

	// Tool: patch_data
	mcp.AddTool(server, &mcp.Tool{
		Name:        "patch_data",
		Description: "Merge new data with existing data in the IoT ephemeral value store. This operation MERGES the new parameters with existing data rather than replacing it. You can specify a nested path (e.g., 'living_room/sensors') to organize data hierarchically. Perfect for multiple IoT devices updating different parts of a shared data structure. Each update includes an automatic timestamp.",
	}, c.PatchDataHandler)

	// Tool: download_data
	mcp.AddTool(server, &mcp.Tool{
		Name:        "download_data",
		Description: "Download data from the IoT ephemeral value store using a download key. You can retrieve all data as a JSON object, or specify a parameter path (e.g., 'temp' or 'living_room/temp') to get a specific value. The download key is read-only and cannot be used to modify data. Supports nested parameter paths using '/' separator.",
	}, c.DownloadDataHandler)

	// Tool: delete_data
	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_data",
		Description: "Delete all data associated with an upload key from the IoT ephemeral value store. This permanently removes all stored values for this key. Note that data is automatically deleted after the configured retention period (default: 24 hours), so manual deletion is optional. Requires the upload key (not the download key).",
	}, c.DeleteDataHandler)
}

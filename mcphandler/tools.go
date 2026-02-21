package mcphandler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolDescriptor defines metadata for a registered MCP tool
type ToolDescriptor struct {
	Name        string
	Description string
}

// toolRegistry defines all available tools with their metadata
var toolRegistry = []ToolDescriptor{
	{
		Name:        "generate_key_pair",
		Description: "Generate a new upload/download key pair for the IoT ephemeral value store. The upload key is used to store data (must be kept secret), and the download key is used to retrieve data (can be shared for read-only access). Keys are cryptographically linked - the download key is derived from the upload key using SHA256.",
	},
	{
		Name:        "upload_data",
		Description: "Upload data to the IoT ephemeral value store using an upload key. This operation REPLACES any existing data for this key with the new data. All parameters are stored with an automatic timestamp. Use this for initial data uploads or complete replacements. For merging with existing data, use patch_data instead.",
	},
	{
		Name:        "patch_data",
		Description: "Merge new data with existing data in the IoT ephemeral value store. This operation MERGES the new parameters with existing data rather than replacing it. You can specify a nested path (e.g., 'living_room/sensors') to organize data hierarchically. Perfect for multiple IoT devices updating different parts of a shared data structure. Each update includes an automatic timestamp.",
	},
	{
		Name:        "download_data",
		Description: "Download data from the IoT ephemeral value store using a download key. You can retrieve all data as a JSON object, or specify a parameter path (e.g., 'temp' or 'living_room/temp') to get a specific value. The download key is read-only and cannot be used to modify data. Supports nested parameter paths using '/' separator.",
	},
	{
		Name:        "delete_data",
		Description: "Delete all data associated with an upload key from the IoT ephemeral value store. This permanently removes all stored values for this key. Note that data is automatically deleted after the configured retention period (default: 24 hours), so manual deletion is optional. Requires the upload key (not the download key).",
	},
}

// RegisteredToolNames contains the names of all registered MCP tools
// This is derived from toolRegistry to ensure consistency
var RegisteredToolNames = extractToolNames()

func extractToolNames() []string {
	names := make([]string, len(toolRegistry))
	for i, tool := range toolRegistry {
		names[i] = tool.Name
	}
	return names
}

// GenerateKeyPairInput represents the input for generating a key pair.
// The noop field exists to satisfy schema generators that disallow empty objects.
type GenerateKeyPairInput struct {
	Noop *struct{} `json:"noop,omitempty" jsonschema:"Optional placeholder field. This tool does not require any input parameters."`
}

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

	uploadKey, downloadKey, err := c.DataService.GenerateKeyPair()
	if err != nil {
		c.StatsInstance.IncrementHTTPErrors()
		return nil, nil, err
	}

	result := map[string]interface{}{
		"upload_key":   uploadKey,
		"download_key": downloadKey,
		"upload_url":   fmt.Sprintf("%s/u/%s?param=value", c.ServerHost, uploadKey),
		"download_url": fmt.Sprintf("%s/d/%s/json", c.ServerHost, downloadKey),
		"message":      "Key pair generated successfully. Use the upload key to store data and the download key to retrieve it. The upload key must be kept secret.",
	}

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("error marshaling result: %w", err)
	}

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

	downloadKey, _, err := c.DataService.Upload(params.UploadKey, params.Parameters)
	if err != nil {
		c.StatsInstance.IncrementHTTPErrors()
		return nil, nil, err
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

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("error marshaling result: %w", err)
	}

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

	downloadKey, _, err := c.DataService.Patch(params.UploadKey, params.Path, params.Parameters)
	if err != nil {
		c.StatsInstance.IncrementHTTPErrors()
		return nil, nil, err
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

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("error marshaling result: %w", err)
	}

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

	var result map[string]interface{}

	if params.Parameter == "" {
		// Return all data as JSON
		jsonData, err := c.DataService.DownloadJSON(params.DownloadKey)
		if err != nil {
			c.StatsInstance.IncrementHTTPErrors()
			return nil, nil, err
		}

		var dataMap map[string]interface{}
		if err := json.Unmarshal(jsonData, &dataMap); err != nil {
			c.StatsInstance.IncrementHTTPErrors()
			return nil, nil, fmt.Errorf("error decoding JSON: %w", err)
		}

		c.StatsInstance.IncrementDownloads()

		result = map[string]interface{}{
			"data":    dataMap,
			"message": "Retrieved all data as JSON",
		}
	} else {
		// Retrieve specific parameter
		value, err := c.DataService.DownloadField(params.DownloadKey, params.Parameter)
		if err != nil {
			c.StatsInstance.IncrementHTTPErrors()
			return nil, nil, err
		}

		c.StatsInstance.IncrementDownloads()

		result = map[string]interface{}{
			"data":      value,
			"parameter": params.Parameter,
			"message":   fmt.Sprintf("Retrieved parameter '%s'", params.Parameter),
		}
	}

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("error marshaling result: %w", err)
	}

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

	_, err := c.DataService.Delete(params.UploadKey)
	if err != nil {
		c.StatsInstance.IncrementHTTPErrors()
		return nil, nil, err
	}

	result := map[string]interface{}{
		"message": "Data deleted successfully",
		"success": true,
	}

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("error marshaling result: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resultJSON)},
		},
	}, nil, nil
}

// RegisterTools registers all MCP tools with the server
func (c Config) RegisterTools(server *mcp.Server) {
	// Tool: generate_key_pair
	tool := getToolByName("generate_key_pair")
	mcp.AddTool(server, &mcp.Tool{
		Name:        tool.Name,
		Description: tool.Description,
	}, c.GenerateKeyPairHandler)

	// Tool: upload_data
	tool = getToolByName("upload_data")
	mcp.AddTool(server, &mcp.Tool{
		Name:        tool.Name,
		Description: tool.Description,
	}, c.UploadDataHandler)

	// Tool: patch_data
	tool = getToolByName("patch_data")
	mcp.AddTool(server, &mcp.Tool{
		Name:        tool.Name,
		Description: tool.Description,
	}, c.PatchDataHandler)

	// Tool: download_data
	tool = getToolByName("download_data")
	mcp.AddTool(server, &mcp.Tool{
		Name:        tool.Name,
		Description: tool.Description,
	}, c.DownloadDataHandler)

	// Tool: delete_data
	tool = getToolByName("delete_data")
	mcp.AddTool(server, &mcp.Tool{
		Name:        tool.Name,
		Description: tool.Description,
	}, c.DeleteDataHandler)
}

// getToolByName retrieves tool metadata by name
func getToolByName(name string) *ToolDescriptor {
	for i, tool := range toolRegistry {
		if tool.Name == name {
			return &toolRegistry[i]
		}
	}
	return nil // Should not happen if toolRegistry is properly maintained
}

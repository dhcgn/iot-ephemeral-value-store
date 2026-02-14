package mcphandler

import (
	"encoding/json"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPServer wraps the MCP server for HTTP handling
type MCPServer struct {
	config  Config
	server  *mcp.Server
	handler http.Handler
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(config Config) (*MCPServer, error) {
	// Use provided version or default to "dev" if empty
	version := config.Version
	if version == "" {
		version = "dev"
	}

	// Create MCP server with implementation and capabilities
	implementation := &mcp.Implementation{
		Name:    "iot-ephemeral-value-store",
		Version: version,
	}

	server := mcp.NewServer(implementation, nil)

	// Register all tools
	config.RegisterTools(server)

	// Create the streamable HTTP handler
	handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
		return server
	}, nil)

	return &MCPServer{
		config:  config,
		server:  server,
		handler: handler,
	}, nil
}

// ServeHTTP handles HTTP requests to the MCP endpoint
func (m *MCPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Use the streamable HTTP handler for POST requests
	if r.Method == http.MethodPost {
		m.handler.ServeHTTP(w, r)
		return
	}

	// Return MCP server information for GET requests
	if r.Method == http.MethodGet {
		m.handleInfoRequest(w, r)
		return
	}

	http.Error(w, "Method not allowed. Use POST for MCP requests or GET for server information.", http.StatusMethodNotAllowed)
}

// handleInfoRequest returns server information for GET requests
func (m *MCPServer) handleInfoRequest(w http.ResponseWriter, r *http.Request) {
	// Use provided version or default
	version := m.config.Version
	if version == "" {
		version = "dev"
	}

	w.Header().Set("Content-Type", "application/json")
	info := map[string]interface{}{
		"protocol": "Model Context Protocol (MCP)",
		"server": map[string]string{
			"name":    "iot-ephemeral-value-store",
			"version": version,
		},
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{
				"available": []string{
					"generate_key_pair",
					"upload_data",
					"patch_data",
					"download_data",
					"delete_data",
				},
			},
		},
		"transport": "HTTP Streamable (SSE)",
		"usage":     "Send POST requests to interact with the server. See documentation at /llm.txt",
	}
	if err := json.NewEncoder(w).Encode(info); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

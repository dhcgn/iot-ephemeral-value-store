package mcphandler

import (
	"github.com/dhcgn/iot-ephemeral-value-store/stats"
	"github.com/dhcgn/iot-ephemeral-value-store/storage"
)

// Config holds the dependencies for MCP handlers
type Config struct {
	StorageInstance storage.Storage
	StatsInstance   *stats.Stats
	ServerHost      string // The base URL of the server for generating links
	Version         string // Application version for MCP server
}

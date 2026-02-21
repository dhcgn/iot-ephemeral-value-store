package mcphandler

import (
	"github.com/dhcgn/iot-ephemeral-value-store/data"
	"github.com/dhcgn/iot-ephemeral-value-store/stats"
)

// Config holds the dependencies for MCP handlers
type Config struct {
	DataService *data.Service
	StatsInstance   *stats.Stats
	ServerHost      string // The base URL of the server for generating links
	Version         string // Application version for MCP server
}

package mcphandler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhcgn/iot-ephemeral-value-store/data"
	"github.com/dhcgn/iot-ephemeral-value-store/stats"
	"github.com/dhcgn/iot-ephemeral-value-store/storage"
)

func newServerTestConfig(version string) Config {
	si := storage.NewInMemoryStorage()
	return Config{
		DataService:   &data.Service{StorageInstance: si},
		StatsInstance: stats.NewStats(),
		ServerHost:    "http://localhost:8080",
		Version:       version,
	}
}

func Test_NewMCPServer(t *testing.T) {
	t.Run("creates server with provided version", func(t *testing.T) {
		cfg := newServerTestConfig("1.2.3")
		srv, err := NewMCPServer(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if srv == nil {
			t.Fatal("expected non-nil server")
		}
		if srv.config.Version != "1.2.3" {
			t.Errorf("expected version 1.2.3, got %q", srv.config.Version)
		}
	})

	t.Run("creates server with empty version", func(t *testing.T) {
		cfg := newServerTestConfig("")
		srv, err := NewMCPServer(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if srv == nil {
			t.Fatal("expected non-nil server")
		}
	})
}

func Test_ServeHTTP_GET(t *testing.T) {
	cfg := newServerTestConfig("2.0.0")
	srv, err := NewMCPServer(cfg)
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var info map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &info); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	serverInfo, ok := info["server"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'server' field in response")
	}
	if serverInfo["version"] != "2.0.0" {
		t.Errorf("expected version 2.0.0, got %v", serverInfo["version"])
	}
}

func Test_ServeHTTP_MethodNotAllowed(t *testing.T) {
	cfg := newServerTestConfig("1.0.0")
	srv, err := NewMCPServer(cfg)
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	for _, method := range []string{http.MethodPut, http.MethodDelete, http.MethodPatch} {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/mcp", nil)
			w := httptest.NewRecorder()
			srv.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected status 405, got %d", w.Code)
			}
		})
	}
}

func Test_handleInfoRequest_defaultVersion(t *testing.T) {
	cfg := newServerTestConfig("")
	srv, err := NewMCPServer(cfg)
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	w := httptest.NewRecorder()
	srv.handleInfoRequest(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var info map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &info); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	serverInfo, ok := info["server"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'server' field in response")
	}
	if serverInfo["version"] != "dev" {
		t.Errorf("expected version 'dev' when empty, got %v", serverInfo["version"])
	}
}

func Test_handleInfoRequest_registeredTools(t *testing.T) {
	cfg := newServerTestConfig("1.0.0")
	srv, err := NewMCPServer(cfg)
	if err != nil {
		t.Fatalf("failed to create MCP server: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	w := httptest.NewRecorder()
	srv.handleInfoRequest(w, req)

	var info map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &info); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	capabilities, ok := info["capabilities"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'capabilities' field in response")
	}
	tools, ok := capabilities["tools"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'tools' field in capabilities")
	}
	available, ok := tools["available"].([]interface{})
	if !ok {
		t.Fatal("expected 'available' field in tools")
	}
	if len(available) != len(RegisteredToolNames) {
		t.Errorf("expected %d tools, got %d", len(RegisteredToolNames), len(available))
	}
}

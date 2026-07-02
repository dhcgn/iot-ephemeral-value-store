package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dhcgn/iot-ephemeral-value-store/storage"
)

func TestMainFunction(t *testing.T) {
	// Set up command line arguments
	os.Args = []string{"cmd", "-persist-values-for=1h", "-store=./testdata", "-port=8081"}

	// Mock functions to avoid actual server start and storage creation
	createStorage = func(storePath string, persistDuration time.Duration) storage.StorageInstance {
		return storage.NewInMemoryStorage()
	}
	listenAndServe = func(srv *http.Server) {
		// Create a test server
		ts := httptest.NewServer(srv.Handler)
		defer ts.Close()

		// Perform a test request
		resp, err := http.Get(ts.URL + "/")
		if err != nil {
			t.Fatalf("Failed to perform test request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status OK, got %v", resp.Status)
		}
	}

	// Call the main function
	main()
}

func TestParseTrustedProxies(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{
			name: "Empty input",
			raw:  "",
			want: nil,
		},
		{
			name: "CIDR input",
			raw:  "172.19.0.0/16",
			want: []string{"172.19.0.0/16"},
		},
		{
			name: "IPv4 and IPv6 addresses",
			raw:  "192.0.2.10, 2001:db8::1",
			want: []string{"192.0.2.10/32", "2001:db8::1/128"},
		},
		{
			name: "Whitespace and invalid entries",
			raw:  " 172.19.0.0/16, invalid , 192.0.2.4 , , 2001:db8::2 ",
			want: []string{"172.19.0.0/16", "192.0.2.4/32", "2001:db8::2/128"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTrustedProxies(tt.raw)
			if len(got) != len(tt.want) {
				t.Fatalf("Expected %d networks, got %d", len(tt.want), len(got))
			}

			for i, network := range got {
				if network.String() != tt.want[i] {
					t.Fatalf("Expected network %q at index %d, got %q", tt.want[i], i, network.String())
				}
			}
		})
	}
}

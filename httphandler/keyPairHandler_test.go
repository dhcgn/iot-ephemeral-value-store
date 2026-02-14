package httphandler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dhcgn/iot-ephemeral-value-store/domain"
	"github.com/dhcgn/iot-ephemeral-value-store/stats"
	"github.com/dhcgn/iot-ephemeral-value-store/storage"
)

func Test_KeyPairHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name                   string
		c                      Config
		args                   args
		expectedStatus         int
		expectedHTTPErrorCount int
	}{
		{
			name: "KeyPairHandler - successful generation",
			c: Config{
				StatsInstance:   stats.NewStats(),
				StorageInstance: storage.NewInMemoryStorage(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/keypair", nil),
			},
			expectedStatus:         http.StatusOK,
			expectedHTTPErrorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.KeyPairHandler(tt.args.w, tt.args.r)

			resp := tt.args.w.(*httptest.ResponseRecorder)
			if resp.Code != tt.expectedStatus {
				t.Errorf("KeyPairHandler returned wrong status code: got \"%v\" want \"%v\"", resp.Code, tt.expectedStatus)
			}

			if tt.c.StatsInstance.GetCurrentStats().HTTPErrorCount != tt.expectedHTTPErrorCount {
				t.Errorf("KeyPairHandler did not increment HTTPErrorCount correctly: got \"%v\" want \"%v\"", tt.c.StatsInstance.GetCurrentStats().HTTPErrorCount, tt.expectedHTTPErrorCount)
			}

			var response map[string]string
			err := json.Unmarshal(resp.Body.Bytes(), &response)
			if err != nil {
				t.Errorf("Failed to unmarshal response body: %v", err)
			}

			uploadKey, ok := response["upload-key"]
			if !ok {
				t.Errorf("Response does not contain upload-key")
			}
			if len(uploadKey) == 0 {
				t.Errorf("Upload key is empty")
			}

			downloadKey, ok := response["download-key"]
			if !ok {
				t.Errorf("Response does not contain download-key")
			}
			if len(downloadKey) == 0 {
				t.Errorf("Download key is empty")
			}

			// Check that upload key has the u_ prefix
			if !strings.HasPrefix(uploadKey, domain.UploadKeyPrefix) {
				t.Errorf("Upload key does not have the expected prefix 'u_': %s", uploadKey)
			}

			// Check that download key has the d_ prefix
			if !strings.HasPrefix(downloadKey, domain.DownloadKeyPrefix) {
				t.Errorf("Download key does not have the expected prefix 'd_': %s", downloadKey)
			}

			// Check if download key is derived correctly from upload key
			derivedDownloadKey, err := domain.DeriveDownloadKey(uploadKey)
			if err != nil {
				t.Errorf("Failed to derive download key: %v", err)
			}
			// Add the prefix to the derived key for comparison
			derivedDownloadKeyWithPrefix := domain.AddDownloadPrefix(derivedDownloadKey)
			if derivedDownloadKeyWithPrefix != downloadKey {
				t.Errorf("Derived download key does not match: got \"%v\" want \"%v\"", derivedDownloadKeyWithPrefix, downloadKey)
			}
		})
	}
}

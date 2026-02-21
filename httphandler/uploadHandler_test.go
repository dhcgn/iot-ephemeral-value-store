package httphandler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dhcgn/iot-ephemeral-value-store/data"
	"github.com/dhcgn/iot-ephemeral-value-store/stats"
	"github.com/dhcgn/iot-ephemeral-value-store/storage"
	"github.com/gorilla/mux"
)

func Test_UploadHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name                   string
		c                      Config
		args                   args
		expectedStatus         int
		expectedBodyContains   []string
		expectedHTTPErrorCount int
		expectedUploadCount    int
	}{
		{
			name: "UploadHandler - invalid upload key",
			c: func() Config {
				si := storage.NewInMemoryStorage()
				return Config{
					StatsInstance: stats.NewStats(),
					DataService:   &data.Service{StorageInstance: si},
				}
			}(),
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, _ := http.NewRequest("POST", "/upload/invalidUploadKey?param1=value1", nil)
					vars := map[string]string{
						"uploadKey": "invalidUploadKey",
					}
					return mux.SetURLVars(req, vars)
				}(),
			},
			expectedStatus:         http.StatusBadRequest,
			expectedBodyContains:   []string{"uploadKey must be a 256 bit hex string"},
			expectedHTTPErrorCount: 1,
			expectedUploadCount:    0,
		},
		{
			name: "UploadHandler - valid upload key",
			c: func() Config {
				si := storage.NewInMemoryStorage()
				return Config{
					StatsInstance: stats.NewStats(),
					DataService:   &data.Service{StorageInstance: si},
				}
			}(),
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, _ := http.NewRequest("POST", "/upload/7790e6a7c72e97c2493334f7b22ffbaa2a41fc53a95268a4fbb45a9c34d9c5d1?param1=value1", nil)
					vars := map[string]string{
						"uploadKey": "7790e6a7c72e97c2493334f7b22ffbaa2a41fc53a95268a4fbb45a9c34d9c5d1",
					}
					return mux.SetURLVars(req, vars)
				}(),
			},
			expectedStatus:         http.StatusOK,
			expectedBodyContains:   []string{"Data uploaded successfully", "download_url", "parameter_urls"},
			expectedHTTPErrorCount: 0,
			expectedUploadCount:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.UploadHandler(tt.args.w, tt.args.r)

			resp := tt.args.w.(*httptest.ResponseRecorder)
			if resp.Code != tt.expectedStatus {
				t.Errorf("UploadHandler returned wrong status code: got \"%v\" want \"%v\"", resp.Code, tt.expectedStatus)
			}

			for _, expectedString := range tt.expectedBodyContains {
				b := resp.Body.String()
				if !contains(b, expectedString) {
					t.Errorf("UploadHandler response does not contain expected string: \"%v\", got \"%v\"", expectedString, b)
				}
			}

			if tt.c.StatsInstance.GetCurrentStats().HTTPErrorCount != tt.expectedHTTPErrorCount {
				t.Errorf("UploadHandler did not increment HTTPErrorCount correctly: got \"%v\" want \"%v\"", tt.c.StatsInstance.GetCurrentStats().HTTPErrorCount, tt.expectedHTTPErrorCount)
			}

			if tt.c.StatsInstance.GetCurrentStats().UploadCount != tt.expectedUploadCount {
				t.Errorf("UploadHandler did not increment UploadCount correctly: got \"%v\" want \"%v\"", tt.c.StatsInstance.GetCurrentStats().UploadCount, tt.expectedUploadCount)
			}
		})
	}
}

func Test_UploadAndPatchHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name                   string
		c                      Config
		args                   args
		expectedStatus         int
		expectedBodyContains   []string
		expectedHTTPErrorCount int
		expectedUploadCount    int
	}{
		{
			name: "UploadAndPatchHandler - invalid upload key",
			c: func() Config {
				si := storage.NewInMemoryStorage()
				return Config{
					StatsInstance: stats.NewStats(),
					DataService:   &data.Service{StorageInstance: si},
				}
			}(),
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, _ := http.NewRequest("PATCH", "/upload/invalidUploadKey/param?value=newvalue", nil)
					vars := map[string]string{
						"uploadKey": "invalidUploadKey",
						"param":     "param",
					}
					return mux.SetURLVars(req, vars)
				}(),
			},
			expectedStatus:         http.StatusBadRequest,
			expectedBodyContains:   []string{"uploadKey must be a 256 bit hex string"},
			expectedHTTPErrorCount: 1,
			expectedUploadCount:    0,
		},
		{
			name: "UploadAndPatchHandler - valid upload key",
			c: func() Config {
				si := storage.NewInMemoryStorage()
				return Config{
					StatsInstance: stats.NewStats(),
					DataService:   &data.Service{StorageInstance: si},
				}
			}(),
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, _ := http.NewRequest("PATCH", "/upload/0143a8a24c3b364ce4df085579601d9f2408f5e93f851078b3f5e4088eb13220/param?value=newvalue", nil)
					vars := map[string]string{
						"uploadKey": "0143a8a24c3b364ce4df085579601d9f2408f5e93f851078b3f5e4088eb13220",
						"param":     "param",
					}
					return mux.SetURLVars(req, vars)
				}(),
			},
			expectedStatus:         http.StatusOK,
			expectedBodyContains:   []string{"Data uploaded successfully", "download_url", "parameter_urls"},
			expectedHTTPErrorCount: 0,
			expectedUploadCount:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.UploadAndPatchHandler(tt.args.w, tt.args.r)

			resp := tt.args.w.(*httptest.ResponseRecorder)
			if resp.Code != tt.expectedStatus {
				t.Errorf("UploadAndPatchHandler returned wrong status code: got \"%v\" want \"%v\"", resp.Code, tt.expectedStatus)
			}

			for _, expectedString := range tt.expectedBodyContains {
				if !contains(resp.Body.String(), expectedString) {
					t.Errorf("UploadAndPatchHandler response does not contain expected string: \"%v\", got \"%v\"", expectedString, resp.Body.String())
				}
			}

			if tt.c.StatsInstance.GetCurrentStats().HTTPErrorCount != tt.expectedHTTPErrorCount {
				t.Errorf("UploadAndPatchHandler did not increment HTTPErrorCount correctly: got \"%v\" want \"%v\"", tt.c.StatsInstance.GetCurrentStats().HTTPErrorCount, tt.expectedHTTPErrorCount)
			}

			if tt.c.StatsInstance.GetCurrentStats().UploadCount != tt.expectedUploadCount {
				t.Errorf("UploadAndPatchHandler did not increment UploadCount correctly: got \"%v\" want \"%v\"", tt.c.StatsInstance.GetCurrentStats().UploadCount, tt.expectedUploadCount)
			}
		})
	}
}

// Helper function to check if a string contains another string
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

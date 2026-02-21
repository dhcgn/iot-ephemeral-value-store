package httphandler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhcgn/iot-ephemeral-value-store/data"
	"github.com/dhcgn/iot-ephemeral-value-store/stats"
	"github.com/dhcgn/iot-ephemeral-value-store/storage"
	"github.com/gorilla/mux"
)

func Test_DeleteHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name                   string
		c                      Config
		args                   args
		expectedStatus         int
		expectedBody           string
		expectedHTTPErrorCount int
	}{
		{
			name: "DeleteHandler - invalid upload key",
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
					req, _ := http.NewRequest("DELETE", "/delete/invalidUploadKey", nil)
					vars := map[string]string{
						"uploadKey": "invalidUploadKey",
					}
					return mux.SetURLVars(req, vars)
				}(),
			},
			expectedStatus:         http.StatusInternalServerError,
			expectedBody:           "invalid upload key: uploadKey must be a 256 bit hex string\n",
			expectedHTTPErrorCount: 1,
		},
		{
			name: "DeleteHandler - unknown upload key",
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
					req, _ := http.NewRequest("DELETE", "/delete/0143a8a24c3b364ce4df085579601d9f2408f5e93f851078b3f5e4088eb13220", nil)
					vars := map[string]string{
						"uploadKey": "0143a8a24c3b364ce4df085579601d9f2408f5e93f851078b3f5e4088eb13220",
					}
					return mux.SetURLVars(req, vars)
				}(),
			},
			expectedStatus:         http.StatusOK,
			expectedBody:           "OK\n",
			expectedHTTPErrorCount: 0,
		},
		{
			name: "DeleteHandler - known upload key",
			c: func() Config {
				si := storage.NewInMemoryStorage()
				_ = si.Store("7790e6a7c72e97c2493334f7b22ffbaa2a41fc53a95268a4fbb45a9c34d9c5d1", map[string]interface{}{"key": "value"})
				return Config{
					StatsInstance: stats.NewStats(),
					DataService:   &data.Service{StorageInstance: si},
				}
			}(),
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, _ := http.NewRequest("DELETE", "/delete/f3749e7288bac3cda9a739f3525da4cc883037e57a984046d5f42d160368078a", nil)
					vars := map[string]string{
						"uploadKey": "f3749e7288bac3cda9a739f3525da4cc883037e57a984046d5f42d160368078a",
					}
					return mux.SetURLVars(req, vars)
				}(),
			},
			expectedStatus:         http.StatusOK,
			expectedBody:           "OK\n",
			expectedHTTPErrorCount: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.DeleteHandler(tt.args.w, tt.args.r)

			resp := tt.args.w.(*httptest.ResponseRecorder)
			if resp.Code != tt.expectedStatus {
				t.Errorf("DeleteHandler returned wrong status code: got \"%v\" want \"%v\"", resp.Code, tt.expectedStatus)
			}

			if resp.Body.String() != tt.expectedBody {
				t.Errorf("DeleteHandler returned unexpected body: got \"%v\" want \"%v\"", resp.Body.String(), tt.expectedBody)
			}

			if tt.c.StatsInstance.GetCurrentStats().HTTPErrorCount != tt.expectedHTTPErrorCount {
				t.Errorf("DeleteHandler did not increment HTTPErrorCount: got \"%v\" want \"%v\"", tt.c.StatsInstance.GetCurrentStats().HTTPErrorCount, tt.expectedHTTPErrorCount)
			}
		})
	}
}

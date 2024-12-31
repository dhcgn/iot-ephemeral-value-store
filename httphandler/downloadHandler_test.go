package httphandler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dhcgn/iot-ephemeral-value-store/stats"
	"github.com/dhcgn/iot-ephemeral-value-store/storage"
	"github.com/gorilla/mux"
)

func Test_PlainDownloadHandler(t *testing.T) {
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
		expectedDownloadCount  int
	}{
		{
			name: "PlainDownloadHandler - invalid download key",
			c: Config{
				StatsInstance:   stats.NewStats(),
				StorageInstance: storage.NewInMemoryStorage(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, _ := http.NewRequest("GET", "/download/invalidDownloadKey/param", nil)
					vars := map[string]string{
						"downloadKey": "invalidDownloadKey",
						"param":       "param",
					}
					return mux.SetURLVars(req, vars)
				}(),
			},
			expectedStatus:         http.StatusNotFound,
			expectedBody:           "Invalid download key or database error\n",
			expectedHTTPErrorCount: 1,
			expectedDownloadCount:  0,
		},
		{
			name: "PlainDownloadHandler - valid download key, invalid param",
			c: Config{
				StatsInstance: stats.NewStats(),
				StorageInstance: func() storage.Storage {
					s := storage.NewInMemoryStorage()
					data := map[string]interface{}{"key": "value"}
					s.Store("validKey", data)
					return s
				}(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, _ := http.NewRequest("GET", "/download/validKey/invalidParam", nil)
					vars := map[string]string{
						"downloadKey": "validKey",
						"param":       "invalidParam",
					}
					return mux.SetURLVars(req, vars)
				}(),
			},
			expectedStatus:         http.StatusNotFound,
			expectedBody:           "Parameter not found\n",
			expectedHTTPErrorCount: 1,
			expectedDownloadCount:  0,
		},
		{
			name: "PlainDownloadHandler - valid download key and param",
			c: Config{
				StatsInstance: stats.NewStats(),
				StorageInstance: func() storage.Storage {
					s := storage.NewInMemoryStorage()
					data := map[string]interface{}{"key": "value"}
					s.Store("validKey", data)
					return s
				}(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, _ := http.NewRequest("GET", "/download/validKey/key", nil)
					vars := map[string]string{
						"downloadKey": "validKey",
						"param":       "key",
					}
					return mux.SetURLVars(req, vars)
				}(),
			},
			expectedStatus:         http.StatusOK,
			expectedBody:           "value\n",
			expectedHTTPErrorCount: 0,
			expectedDownloadCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.DownloadPlainHandler(tt.args.w, tt.args.r)

			resp := tt.args.w.(*httptest.ResponseRecorder)
			if resp.Code != tt.expectedStatus {
				t.Errorf("PlainDownloadHandler returned wrong status code: got \"%v\" want \"%v\"", resp.Code, tt.expectedStatus)
			}

			if resp.Body.String() != tt.expectedBody {
				t.Errorf("PlainDownloadHandler returned unexpected body: got \"%v\" want \"%v\"", resp.Body.String(), tt.expectedBody)
			}

			if tt.c.StatsInstance.GetCurrentStats().HTTPErrorCount != tt.expectedHTTPErrorCount {
				t.Errorf("PlainDownloadHandler did not increment HTTPErrorCount correctly: got \"%v\" want \"%v\"", tt.c.StatsInstance.GetCurrentStats().HTTPErrorCount, tt.expectedHTTPErrorCount)
			}

			if tt.c.StatsInstance.GetCurrentStats().DownloadCount != tt.expectedDownloadCount {
				t.Errorf("PlainDownloadHandler did not increment DownloadCount correctly: got \"%v\" want \"%v\"", tt.c.StatsInstance.GetCurrentStats().DownloadCount, tt.expectedDownloadCount)
			}
		})
	}
}

func Test_DownloadHandler(t *testing.T) {
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
		expectedDownloadCount  int
	}{
		{
			name: "DownloadHandler - invalid download key",
			c: Config{
				StatsInstance:   stats.NewStats(),
				StorageInstance: storage.NewInMemoryStorage(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, _ := http.NewRequest("GET", "/download/invalidDownloadKey", nil)
					vars := map[string]string{
						"downloadKey": "invalidDownloadKey",
					}
					return mux.SetURLVars(req, vars)
				}(),
			},
			expectedStatus:         http.StatusNotFound,
			expectedBody:           "Invalid download key or database error\n",
			expectedHTTPErrorCount: 1,
			expectedDownloadCount:  0,
		},
		{
			name: "DownloadHandler - valid download key",
			c: Config{
				StatsInstance: stats.NewStats(),
				StorageInstance: func() storage.Storage {
					s := storage.NewInMemoryStorage()
					data := map[string]interface{}{"key": "value"}
					s.Store("validKey", data)
					return s
				}(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, _ := http.NewRequest("GET", "/download/validKey", nil)
					vars := map[string]string{
						"downloadKey": "validKey",
					}
					return mux.SetURLVars(req, vars)
				}(),
			},
			expectedStatus:         http.StatusOK,
			expectedBody:           `{"key":"value"}`,
			expectedHTTPErrorCount: 0,
			expectedDownloadCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.DownloadJsonHandler(tt.args.w, tt.args.r)

			resp := tt.args.w.(*httptest.ResponseRecorder)
			if resp.Code != tt.expectedStatus {
				t.Errorf("DownloadHandler returned wrong status code: got \"%v\" want \"%v\"", resp.Code, tt.expectedStatus)
			}

			if resp.Body.String() != tt.expectedBody {
				t.Errorf("DownloadHandler returned unexpected body: got \"%v\" want \"%v\"", resp.Body.String(), tt.expectedBody)
			}

			if tt.c.StatsInstance.GetCurrentStats().HTTPErrorCount != tt.expectedHTTPErrorCount {
				t.Errorf("DownloadHandler did not increment HTTPErrorCount correctly: got \"%v\" want \"%v\"", tt.c.StatsInstance.GetCurrentStats().HTTPErrorCount, tt.expectedHTTPErrorCount)
			}

			if tt.c.StatsInstance.GetCurrentStats().DownloadCount != tt.expectedDownloadCount {
				t.Errorf("DownloadHandler did not increment DownloadCount correctly: got \"%v\" want \"%v\"", tt.c.StatsInstance.GetCurrentStats().DownloadCount, tt.expectedDownloadCount)
			}
		})
	}
}

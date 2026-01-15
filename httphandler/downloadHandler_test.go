package httphandler

import (
	"encoding/base64"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dhcgn/iot-ephemeral-value-store/stats"
	"github.com/dhcgn/iot-ephemeral-value-store/storage"
	"github.com/gorilla/mux"
)

// getTestDownloadTemplate returns a template for testing
func getTestDownloadTemplate() *template.Template {
	tmplStr := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Download Options</title>
</head>
<body>
    <h1>Download Options</h1>
    <div class="section">
        <h2>JSON Format</h2>
        <ul>
            <li><a href="/d/{{.DownloadKey}}/json">/d/{{.DownloadKey}}/json</a></li>
        </ul>
    </div>
    <div class="section">
        <h2>Plain Text Fields</h2>
        <ul>
            {{range .Fields}}
            <li><a href="/d/{{$.DownloadKey}}/plain/{{.URLEncoded}}">/d/{{$.DownloadKey}}/plain/{{.Name}}</a></li>
            {{end}}
        </ul>
    </div>
</body>
</html>`
	tmpl, err := template.New("download").Parse(tmplStr)
	if err != nil {
		panic(err)
	}
	return tmpl
}

func Test_PlainDownloadHandler(t *testing.T) {
	type args struct {
		w          http.ResponseWriter
		r          *http.Request
		base64mode bool
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
		{
			name: "PlainDownloadHandler - error decoding JSON",
			c: Config{
				StatsInstance: stats.NewStats(),
				StorageInstance: func() storage.Storage {
					s := storage.NewInMemoryStorage()
					// Store invalid JSON data
					s.StoreRawForTesting("validKey", []byte(`{"key": "value"`)) // Missing closing brace
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
			expectedStatus:         http.StatusInternalServerError,
			expectedBody:           "Error decoding JSON\n",
			expectedHTTPErrorCount: 1,
			expectedDownloadCount:  0,
		},
		{
			name: "PlainDownloadHandler - invalid parameter path",
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
					req, _ := http.NewRequest("GET", "/download/validKey/key/invalidPath", nil)
					vars := map[string]string{
						"downloadKey": "validKey",
						"param":       "key/invalidPath",
					}
					return mux.SetURLVars(req, vars)
				}(),
			},
			expectedStatus:         http.StatusBadRequest,
			expectedBody:           "Invalid parameter path\n",
			expectedHTTPErrorCount: 1,
			expectedDownloadCount:  0,
		},
		{
			name: "PlainDownloadHandler - error decoding base64url",
			c: Config{
				StatsInstance: stats.NewStats(),
				StorageInstance: func() storage.Storage {
					s := storage.NewInMemoryStorage()
					data := map[string]interface{}{"key": "invalid_base64"}
					s.Store("validKey", data)
					return s
				}(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, _ := http.NewRequest("GET", "/download/validKey/plain-from-base64url/key", nil)
					vars := map[string]string{
						"downloadKey": "validKey",
						"param":       "key",
					}
					return mux.SetURLVars(req, vars)
				}(),
				base64mode: true,
			},
			expectedStatus:         http.StatusInternalServerError,
			expectedBody:           "Error decoding base64url\n",
			expectedHTTPErrorCount: 1,
			expectedDownloadCount:  0,
		},
		{
			name: "PlainDownloadHandler - decoding base64url",
			c: Config{
				StatsInstance: stats.NewStats(),
				StorageInstance: func() storage.Storage {
					s := storage.NewInMemoryStorage()
					base64string := base64.URLEncoding.EncodeToString([]byte("Hallo Welt!"))
					data := map[string]interface{}{"key": base64string}
					s.Store("validKey", data)
					return s
				}(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, _ := http.NewRequest("GET", "/download/validKey/plain-from-base64url/key", nil)
					vars := map[string]string{
						"downloadKey": "validKey",
						"param":       "key",
					}
					return mux.SetURLVars(req, vars)
				}(),
				base64mode: true,
			},
			expectedStatus:         http.StatusOK,
			expectedBody:           "Hallo Welt!\n",
			expectedHTTPErrorCount: 0,
			expectedDownloadCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.base64mode {
				tt.c.DownloadBase64Handler(tt.args.w, tt.args.r)
			} else {
				tt.c.DownloadPlainHandler(tt.args.w, tt.args.r)
			}

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

func Test_DownloadRootHandler(t *testing.T) {
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
		expectedDownloadCount  int
	}{
		{
			name: "DownloadRootHandler - invalid download key",
			c: Config{
				StatsInstance:    stats.NewStats(),
				StorageInstance:  storage.NewInMemoryStorage(),
				DownloadTemplate: getTestDownloadTemplate(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, _ := http.NewRequest("GET", "/d/invalidDownloadKey/", nil)
					vars := map[string]string{
						"downloadKey": "invalidDownloadKey",
					}
					return mux.SetURLVars(req, vars)
				}(),
			},
			expectedStatus:         http.StatusNotFound,
			expectedBodyContains:   []string{"Invalid download key or database error"},
			expectedHTTPErrorCount: 1,
			expectedDownloadCount:  0,
		},
		{
			name: "DownloadRootHandler - valid download key with single field",
			c: Config{
				StatsInstance: stats.NewStats(),
				StorageInstance: func() storage.Storage {
					s := storage.NewInMemoryStorage()
					data := map[string]interface{}{"name": "value"}
					s.Store("validKey", data)
					return s
				}(),
				DownloadTemplate: getTestDownloadTemplate(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, _ := http.NewRequest("GET", "/d/validKey/", nil)
					vars := map[string]string{
						"downloadKey": "validKey",
					}
					return mux.SetURLVars(req, vars)
				}(),
			},
			expectedStatus: http.StatusOK,
			expectedBodyContains: []string{
				"Download Options",
				"/d/validKey/json",
				"/d/validKey/plain/name",
			},
			expectedHTTPErrorCount: 0,
			expectedDownloadCount:  1,
		},
		{
			name: "DownloadRootHandler - valid download key with multiple fields",
			c: Config{
				StatsInstance: stats.NewStats(),
				StorageInstance: func() storage.Storage {
					s := storage.NewInMemoryStorage()
					data := map[string]interface{}{
						"name":   "value",
						"temp":   "25",
						"status": "ok",
					}
					s.Store("validKey", data)
					return s
				}(),
				DownloadTemplate: getTestDownloadTemplate(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, _ := http.NewRequest("GET", "/d/validKey/", nil)
					vars := map[string]string{
						"downloadKey": "validKey",
					}
					return mux.SetURLVars(req, vars)
				}(),
			},
			expectedStatus: http.StatusOK,
			expectedBodyContains: []string{
				"Download Options",
				"/d/validKey/json",
				"/d/validKey/plain/name",
				"/d/validKey/plain/temp",
				"/d/validKey/plain/status",
			},
			expectedHTTPErrorCount: 0,
			expectedDownloadCount:  1,
		},
		{
			name: "DownloadRootHandler - HTML escaping for XSS prevention",
			c: Config{
				StatsInstance: stats.NewStats(),
				StorageInstance: func() storage.Storage {
					s := storage.NewInMemoryStorage()
					data := map[string]interface{}{
						"<script>alert('xss')</script>": "value",
						"normal_field":                  "value",
					}
					s.Store("validKey", data)
					return s
				}(),
				DownloadTemplate: getTestDownloadTemplate(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, _ := http.NewRequest("GET", "/d/validKey/", nil)
					vars := map[string]string{
						"downloadKey": "validKey",
					}
					return mux.SetURLVars(req, vars)
				}(),
			},
			expectedStatus: http.StatusOK,
			expectedBodyContains: []string{
				"Download Options",
				"/d/validKey/json",
				"&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;", // Escaped HTML
				"/d/validKey/plain/normal_field",
			},
			expectedHTTPErrorCount: 0,
			expectedDownloadCount:  1,
		},
		{
			name: "DownloadRootHandler - error decoding JSON",
			c: Config{
				StatsInstance: stats.NewStats(),
				StorageInstance: func() storage.Storage {
					s := storage.NewInMemoryStorage()
					// Store invalid JSON data
					s.StoreRawForTesting("validKey", []byte(`{"key": "value"`)) // Missing closing brace
					return s
				}(),
				DownloadTemplate: getTestDownloadTemplate(),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req, _ := http.NewRequest("GET", "/d/validKey/", nil)
					vars := map[string]string{
						"downloadKey": "validKey",
					}
					return mux.SetURLVars(req, vars)
				}(),
			},
			expectedStatus:         http.StatusInternalServerError,
			expectedBodyContains:   []string{"Error decoding JSON"},
			expectedHTTPErrorCount: 1,
			expectedDownloadCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.DownloadRootHandler(tt.args.w, tt.args.r)

			resp := tt.args.w.(*httptest.ResponseRecorder)
			if resp.Code != tt.expectedStatus {
				t.Errorf("DownloadRootHandler returned wrong status code: got \"%v\" want \"%v\"", resp.Code, tt.expectedStatus)
			}

			body := resp.Body.String()
			for _, expected := range tt.expectedBodyContains {
				if !strings.Contains(body, expected) {
					t.Errorf("DownloadRootHandler body does not contain expected string: want \"%v\" in \"%v\"", expected, body)
				}
			}

			if tt.c.StatsInstance.GetCurrentStats().HTTPErrorCount != tt.expectedHTTPErrorCount {
				t.Errorf("DownloadRootHandler did not increment HTTPErrorCount correctly: got \"%v\" want \"%v\"", tt.c.StatsInstance.GetCurrentStats().HTTPErrorCount, tt.expectedHTTPErrorCount)
			}

			if tt.c.StatsInstance.GetCurrentStats().DownloadCount != tt.expectedDownloadCount {
				t.Errorf("DownloadRootHandler did not increment DownloadCount correctly: got \"%v\" want \"%v\"", tt.c.StatsInstance.GetCurrentStats().DownloadCount, tt.expectedDownloadCount)
			}
		})
	}
}


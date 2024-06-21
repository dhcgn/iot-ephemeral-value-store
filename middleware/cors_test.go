package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnableCORS(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkHeaders   bool
		expectedBody   string
	}{
		{
			name:           "GET request",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
			expectedBody:   "OK",
		},
		{
			name:           "POST request",
			method:         "POST",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
			expectedBody:   "OK",
		},
		{
			name:           "OPTIONS request",
			method:         "OPTIONS",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
			expectedBody:   "",
		},
		{
			name:           "Unsupported method",
			method:         "PUT",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
			expectedBody:   "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock handler that the middleware will wrap
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			// Create the middleware
			config := Config{}
			middleware := config.EnableCORS(handler)

			// Create a test request
			req := httptest.NewRequest(tt.method, "/", nil)
			rr := httptest.NewRecorder()

			// Serve the request through the middleware
			middleware.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			// Check the CORS headers
			if tt.checkHeaders {
				headers := rr.Header()
				expectedHeaders := map[string]string{
					"Access-Control-Allow-Origin":  "*",
					"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
					"Access-Control-Allow-Headers": "Content-Type, Authorization",
				}

				for key, value := range expectedHeaders {
					if headers.Get(key) != value {
						t.Errorf("handler returned wrong %s header: got %v want %v",
							key, headers.Get(key), value)
					}
				}
			}

			// Check the response body
			body, _ := io.ReadAll(rr.Body)
			if string(body) != tt.expectedBody {
				t.Errorf("handler returned unexpected body: got %v want %v",
					string(body), tt.expectedBody)
			}
		})
	}
}

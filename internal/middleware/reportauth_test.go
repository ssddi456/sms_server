package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestReportKeyAuth_DisabledWhenKeyEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ReportKeyAuth("", nil))
	r.POST("/sms/webhook", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/sms/webhook", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d want %d", w.Code, http.StatusOK)
	}
}

func TestReportKeyAuth_RejectsMissingOrInvalidKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ReportKeyAuth("my-secret-key", nil))
	r.POST("/sms/webhook", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	tests := []struct {
		name      string
		headers   map[string]string
		wantCode  int
	}{
		{
			name:     "missing key",
			headers:  map[string]string{},
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "invalid key in X-Report-Key",
			headers: map[string]string{
				"X-Report-Key": "wrong-key",
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "invalid key in Authorization",
			headers: map[string]string{
				"Authorization": "Bearer wrong-key",
			},
			wantCode: http.StatusUnauthorized,
		},
		{
			name: "valid key in X-Report-Key",
			headers: map[string]string{
				"X-Report-Key": "my-secret-key",
			},
			wantCode: http.StatusOK,
		},
		{
			name: "valid key in Authorization",
			headers: map[string]string{
				"Authorization": "Bearer my-secret-key",
			},
			wantCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/sms/webhook", nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.wantCode {
				t.Fatalf("status=%d want %d body=%s", w.Code, tc.wantCode, w.Body.String())
			}
		})
	}
}

func TestReportKeysAuth_AllowsRotatedKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ReportKeysAuth([]string{"old-key", "new-key"}, nil))
	r.POST("/sms/webhook", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	tests := []struct {
		name     string
		key      string
		wantCode int
	}{
		{name: "old key allowed", key: "old-key", wantCode: http.StatusOK},
		{name: "new key allowed", key: "new-key", wantCode: http.StatusOK},
		{name: "unknown key denied", key: "bad-key", wantCode: http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/sms/webhook", nil)
			req.Header.Set("X-Report-Key", tc.key)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.wantCode {
				t.Fatalf("status=%d want %d body=%s", w.Code, tc.wantCode, w.Body.String())
			}
		})
	}
}

func TestReportKeyAuth_AllowsReportKeyQueryParam(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ReportKeyAuth("query-key", nil))
	r.POST("/sms/webhook", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/sms/webhook?report_key=query-key", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d want %d body=%s", w.Code, http.StatusOK, w.Body.String())
	}
}

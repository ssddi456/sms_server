package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ssddi456/sms_server/internal/logging"
)

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.Use(RequestLog(logging.NewLogger("info")))
	r.GET("/ping", func(c *gin.Context) {
		if c.GetString("request_id") == "" {
			t.Fatal("request_id missing")
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if got := w.Header().Get("X-Request-ID"); got == "" {
		t.Fatal("X-Request-ID header missing")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d want %d", w.Code, http.StatusOK)
	}
}

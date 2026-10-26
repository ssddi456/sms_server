package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ssddi456/sms_server/internal/logging"
	"github.com/ssddi456/sms_server/internal/store"
)

func TestWebhookHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	st, err := store.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	h := New(st, logging.NewLogger("info"))
	r := gin.New()
	r.POST("/sms/webhook", h.Webhook)

	body := `{"from":"10086","sms":"您的验证码是123456，请在5分钟内使用。","sim_slot":"SIM1(13800138000)","receive_time":"2025-03-15T14:30:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/sms/webhook", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if got := resp["code"]; got != "SUCCESS" {
		t.Fatalf("code=%v want SUCCESS", got)
	}
	data := resp["data"].(map[string]interface{})
	if data["extracted_code"] != "123456" {
		t.Fatalf("extracted_code=%v want 123456", data["extracted_code"])
	}
}

func TestHistoryHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	st, err := store.NewStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	_, err = st.SaveWebhook(store.WebhookRequest{From: "10086", SMS: "验证码: 123456", SimSlot: "SIM1(13800138000)", ReceiveTime: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.SaveWebhook(store.WebhookRequest{From: "10010", SMS: "验证码: 654321", SimSlot: "SIM1(13900139000)", ReceiveTime: time.Now()})
	if err != nil {
		t.Fatal(err)
	}

	h := New(st, logging.NewLogger("info"))
	r := gin.New()
	r.GET("/sms/history", h.History)

	req := httptest.NewRequest(http.MethodGet, "/sms/history?limit=1&page=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["result"] != "ok" {
		t.Fatalf("result=%v want ok", resp["result"])
	}
}

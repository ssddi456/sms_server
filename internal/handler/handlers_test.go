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

type webhookResponse struct {
	Result  string `json:"result"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ID            int    `json:"id"`
		ExtractedCode string `json:"extracted_code"`
		PhoneNumber   string `json:"phone_number"`
	} `json:"data"`
}

type historyResponse struct {
	Result  string `json:"result"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Page  int               `json:"page"`
		Limit int               `json:"limit"`
		Items []store.SMSRecord `json:"items"`
	} `json:"data"`
}

func newTestRouter(t *testing.T) (*gin.Engine, *store.Store) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	st, err := store.NewStore(t.TempDir())
	if err != nil {
		if strings.Contains(err.Error(), "requires cgo") {
			t.Skip("sqlite3 cgo is unavailable in current environment")
		}
		t.Fatal(err)
	}

	h := New(st, logging.NewLogger("info"))
	r := gin.New()
	r.POST("/sms/webhook", h.Webhook)
	r.POST("/v1/sms/receive", h.Webhook)
	r.GET("/sms/history", h.History)
	return r, st
}

func postJSON(t *testing.T, r *gin.Engine, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestWebhookHandler_SmsForwarderPayloadChain(t *testing.T) {
	r, st := newTestRouter(t)
	defer st.Close()

	payload := `{
		"from":"10086",
		"contact_name":"中国移动",
		"phone_area":"86",
		"sms":"您的验证码是123456，请在5分钟内完成验证。",
		"sim_slot":"SIM1(13800138000)",
		"sim_sub_id":"1",
		"device_name":"DemoPhone",
		"receive_time":"2025-03-15T14:30:00Z"
	}`
	webhookResp := postJSON(t, r, "/v1/sms/receive", payload)
	if webhookResp.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", webhookResp.Code, webhookResp.Body.String())
	}

	var webhookBody webhookResponse
	if err := json.Unmarshal(webhookResp.Body.Bytes(), &webhookBody); err != nil {
		t.Fatal(err)
	}
	if webhookBody.Code != "SUCCESS" {
		t.Fatalf("code=%s want SUCCESS", webhookBody.Code)
	}
	if webhookBody.Data.ID <= 0 {
		t.Fatalf("id=%d want > 0", webhookBody.Data.ID)
	}
	if webhookBody.Data.ExtractedCode != "123456" {
		t.Fatalf("extracted_code=%s want 123456", webhookBody.Data.ExtractedCode)
	}
	if webhookBody.Data.PhoneNumber != "13800138000" {
		t.Fatalf("phone_number=%s want 13800138000", webhookBody.Data.PhoneNumber)
	}

	req := httptest.NewRequest(http.MethodGet, "/sms/history?limit=20&page=1", nil)
	historyResp := httptest.NewRecorder()
	r.ServeHTTP(historyResp, req)
	if historyResp.Code != http.StatusOK {
		t.Fatalf("history status=%d body=%s", historyResp.Code, historyResp.Body.String())
	}

	var historyBody historyResponse
	if err := json.Unmarshal(historyResp.Body.Bytes(), &historyBody); err != nil {
		t.Fatal(err)
	}
	if historyBody.Code != "SUCCESS" {
		t.Fatalf("history code=%s want SUCCESS", historyBody.Code)
	}
	if len(historyBody.Data.Items) == 0 {
		t.Fatal("history items is empty")
	}

	var matched *store.SMSRecord
	for i := range historyBody.Data.Items {
		if historyBody.Data.Items[i].ID == webhookBody.Data.ID {
			matched = &historyBody.Data.Items[i]
			break
		}
	}
	if matched == nil {
		t.Fatalf("record id=%d not found in history", webhookBody.Data.ID)
	}
	if matched.From != "10086" || matched.ContactName != "中国移动" || matched.PhoneArea != "86" {
		t.Fatalf("unexpected sender data: %+v", *matched)
	}
	if matched.DeviceName != "DemoPhone" || matched.SimSubID != "1" {
		t.Fatalf("unexpected device/sim data: %+v", *matched)
	}
	if matched.ExtractedCode != "123456" || matched.PhoneNumber != "13800138000" {
		t.Fatalf("unexpected extracted values: %+v", *matched)
	}
}

func TestWebhookHandler_InvalidRequestCases(t *testing.T) {
	r, st := newTestRouter(t)
	defer st.Close()

	testCases := []struct {
		name       string
		body       string
		wantStatus int
		wantCode   string
	}{
		{name: "invalid json", body: `{"from":"10086"`, wantStatus: http.StatusBadRequest, wantCode: "INVALID_JSON"},
		{name: "missing sms", body: `{"from":"10086","sms":"  "}`, wantStatus: http.StatusBadRequest, wantCode: "INVALID_PARAM"},
		{name: "missing from", body: `{"from":" ","sms":"验证码 123456"}`, wantStatus: http.StatusBadRequest, wantCode: "INVALID_PARAM"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := postJSON(t, r, "/sms/webhook", tc.body)
			if w.Code != tc.wantStatus {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}

			var resp webhookResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatal(err)
			}
			if resp.Code != tc.wantCode {
				t.Fatalf("code=%s want %s", resp.Code, tc.wantCode)
			}
		})
	}
}

func TestHistoryHandler_PaginationFallback(t *testing.T) {
	r, st := newTestRouter(t)
	defer st.Close()

	_, err := st.SaveWebhook(store.WebhookRequest{
		From:        "10086",
		ContactName: "A",
		SMS:         "验证码: 123456",
		SimSlot:     "SIM1(13800138000)",
		ReceiveTime: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/sms/history?limit=-1&page=0", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}

	var resp historyResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Limit != 20 {
		t.Fatalf("limit=%d want 20", resp.Data.Limit)
	}
	if resp.Data.Page != 1 {
		t.Fatalf("page=%d want 1", resp.Data.Page)
	}
}

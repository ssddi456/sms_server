package handler

import (
	"encoding/json"
	"io"
	"net/url"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ssddi456/sms_server/internal/logging"
	"github.com/ssddi456/sms_server/internal/store"
)

type Handler struct {
	Store  *store.Store
	Logger *logging.Logger
}

func New(store *store.Store, logger *logging.Logger) *Handler {
	return &Handler{Store: store, Logger: logger}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"result":  "ok",
		"code":    "SUCCESS",
		"message": "ok",
		"data": gin.H{
			"status": "healthy",
		},
	})
}

func (h *Handler) UI(c *gin.Context) {
	c.Redirect(http.StatusTemporaryRedirect, "/sms/assets/app/index.html")
}

func (h *Handler) Webhook(c *gin.Context) {
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logRequest("invalid_body", c, map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusBadRequest, gin.H{
			"result":  "error",
			"code":    "INVALID_BODY",
			"message": "invalid request body",
		})
		return
	}

	var req store.WebhookRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		contentType := c.ContentType()
		allowFormFallback := contentType == "" ||
			contentType == "application/x-www-form-urlencoded" ||
			contentType == "multipart/form-data"
		if !allowFormFallback {
			h.logRequest("invalid_json", c, map[string]interface{}{"error": err.Error()})
			c.JSON(http.StatusBadRequest, gin.H{
				"result":  "error",
				"code":    "INVALID_JSON",
				"message": "invalid request body",
			})
			return
		}

		values, formErr := url.ParseQuery(string(rawBody))
		if formErr != nil {
			h.logRequest("invalid_body", c, map[string]interface{}{
				"json_error": err.Error(),
				"form_error": formErr.Error(),
			})
			c.JSON(http.StatusBadRequest, gin.H{
				"result":  "error",
				"code":    "INVALID_BODY",
				"message": "invalid request body",
			})
			return
		}
		req = normalizeFormWebhook(values)
	}
	if strings.TrimSpace(req.SMS) == "" {
		h.logRequest("invalid_param", c, map[string]interface{}{"error": "sms is required"})
		c.JSON(http.StatusBadRequest, gin.H{
			"result":  "error",
			"code":    "INVALID_PARAM",
			"message": "sms is required",
		})
		return
	}
	if strings.TrimSpace(req.From) == "" {
		h.logRequest("invalid_param", c, map[string]interface{}{"error": "from is required"})
		c.JSON(http.StatusBadRequest, gin.H{
			"result":  "error",
			"code":    "INVALID_PARAM",
			"message": "from is required",
		})
		return
	}

	rec, err := h.Store.SaveWebhook(req)
	if err != nil {
		h.logRequest("store_error", c, map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{
			"result":  "error",
			"code":    "STORE_ERROR",
			"message": "failed to store sms",
		})
		return
	}

	h.Logger.Info("sms_received", map[string]interface{}{
		"request_id":     c.GetString("request_id"),
		"from":           rec.From,
		"phone_number":   maskPhone(rec.PhoneNumber),
		"extracted_code": rec.ExtractedCode,
		"sms_preview":    maskSMS(rec.SMS),
		"receive_time":   rec.ReceiveTime.Format("2006-01-02T15:04:05Z07:00"),
	})

	c.JSON(http.StatusOK, gin.H{
		"result":  "ok",
		"code":    "SUCCESS",
		"message": "短信接收成功",
		"data": gin.H{
			"id":             rec.ID,
			"extracted_code": rec.ExtractedCode,
			"phone_number":   rec.PhoneNumber,
		},
	})
}

func normalizeFormWebhook(values url.Values) store.WebhookRequest {
	req := store.WebhookRequest{
		From:        firstNonEmpty(values.Get("from"), values.Get("sender"), values.Get("phone")),
		ContactName: firstNonEmpty(values.Get("contact_name"), values.Get("contactName")),
		PhoneArea:   firstNonEmpty(values.Get("phone_area"), values.Get("phoneArea")),
		SMS:         firstNonEmpty(values.Get("sms"), values.Get("msg"), values.Get("message"), values.Get("content"), values.Get("text")),
		SimSlot:     firstNonEmpty(values.Get("sim_slot"), values.Get("simSlot")),
		SimSubID:    firstNonEmpty(values.Get("sim_sub_id"), values.Get("simSubId"), values.Get("simSubID")),
		DeviceName:  firstNonEmpty(values.Get("device_name"), values.Get("deviceName")),
	}
	receiveTimeText := firstNonEmpty(values.Get("receive_time"), values.Get("receiveTime"))
	if receiveTimeText != "" {
		if t, err := time.Parse(time.RFC3339, receiveTimeText); err == nil {
			req.ReceiveTime = t
		}
	}
	return req
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func (h *Handler) History(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	records, err := h.Store.History(limit, offset)
	if err != nil {
		h.logRequest("history_error", c, map[string]interface{}{"error": err.Error()})
		c.JSON(http.StatusInternalServerError, gin.H{
			"result":  "error",
			"code":    "QUERY_ERROR",
			"message": "failed to query history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":  "ok",
		"code":    "SUCCESS",
		"message": "ok",
		"data": gin.H{
			"page":  page,
			"limit": limit,
			"items": records,
		},
	})
}

func (h *Handler) logRequest(event string, c *gin.Context, fields map[string]interface{}) {
	if h.Logger == nil {
		return
	}
	payload := map[string]interface{}{
		"request_id": c.GetString("request_id"),
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"event":      event,
	}
	for k, v := range fields {
		payload[k] = v
	}
	h.Logger.Error(event, payload)
}

func maskPhone(phone string) string {
	if phone == "" {
		return ""
	}
	if len(phone) <= 4 {
		return "****"
	}
	if len(phone) <= 7 {
		return phone[:3] + "****"
	}
	return phone[:3] + "****" + phone[len(phone)-2:]
}

func maskSMS(s string) string {
	if len(s) <= 12 {
		return "****"
	}
	return s[:6] + "..." + s[len(s)-4:]
}

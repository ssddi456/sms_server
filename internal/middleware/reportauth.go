package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ssddi456/sms_server/internal/logging"
)

func ReportKeyAuth(requiredKey string, logger *logging.Logger) gin.HandlerFunc {
	return ReportKeysAuth([]string{requiredKey}, logger)
}

func ReportKeysAuth(requiredKeys []string, logger *logging.Logger) gin.HandlerFunc {
	normalizedKeys := normalizeRequiredKeys(requiredKeys)

	return func(c *gin.Context) {
		if len(normalizedKeys) == 0 {
			c.Next()
			return
		}

		providedKey := strings.TrimSpace(c.GetHeader("X-Report-Key"))
		if providedKey == "" {
			authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
			if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				providedKey = strings.TrimSpace(authHeader[7:])
			}
		}
		if providedKey == "" {
			providedKey = strings.TrimSpace(c.Query("report_key"))
		}

		if !isValidReportKey(providedKey, normalizedKeys) {
			if logger != nil {
				logger.Error("report_auth_failed", map[string]interface{}{
					"request_id": c.GetString("request_id"),
					"method":     c.Request.Method,
					"path":       c.Request.URL.Path,
					"client_ip":  c.ClientIP(),
				})
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"result":  "error",
				"code":    "UNAUTHORIZED",
				"message": "invalid report key",
			})
			return
		}

		c.Next()
	}
}

func normalizeRequiredKeys(requiredKeys []string) []string {
	normalized := make([]string, 0, len(requiredKeys))
	seen := make(map[string]struct{})
	for _, key := range requiredKeys {
		trimmed := strings.TrimSpace(key)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

func isValidReportKey(providedKey string, requiredKeys []string) bool {
	for _, requiredKey := range requiredKeys {
		if subtle.ConstantTimeCompare([]byte(providedKey), []byte(requiredKey)) == 1 {
			return true
		}
	}
	return false
}

package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/ssddi456/sms_server/internal/config"
	"github.com/ssddi456/sms_server/internal/handler"
	"github.com/ssddi456/sms_server/internal/logging"
	"github.com/ssddi456/sms_server/internal/middleware"
	"github.com/ssddi456/sms_server/internal/store"
)

func main() {
	cfg := config.Load()
	gin.SetMode(cfg.GinMode)
	logger := logging.NewLogger(cfg.LogLevel)

	storeInstance, err := store.NewStore(cfg.DataDir)
	if err != nil {
		log.Fatalf("init store: %v", err)
	}
	defer storeInstance.Close()

	h := handler.New(storeInstance, logger)
	r := gin.New()
	r.LoadHTMLFiles("frontend/templates/ui.html")
	r.Static("/sms/assets", "frontend/static")
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLog(logger))

	reportAuth := middleware.ReportKeysAuth(cfg.ReportAuthKeys, logger)

	smsGroup := r.Group("/sms")
	smsGroup.GET("/ui", h.UI)
	smsGroup.GET("/healthz", h.Health)
	smsGroup.POST("/webhook", reportAuth, h.Webhook)
	smsGroup.GET("/history", reportAuth, h.History)

	legacyGroup := r.Group("/v1/sms")
	legacyGroup.GET("/healthz", h.Health)
	legacyGroup.POST("/receive", reportAuth, h.Webhook)
	legacyGroup.GET("/history", reportAuth, h.History)

	logger.Info("server_start", map[string]interface{}{
		"host":                  cfg.Host,
		"port":                  cfg.Port,
		"data_dir":              cfg.DataDir,
		"report_auth_enabled":   len(cfg.ReportAuthKeys) > 0,
		"report_auth_keys_count": len(cfg.ReportAuthKeys),
	})
	fmt.Printf("listening on %s\n", cfg.ListenAddr)
	if err := r.Run(cfg.ListenAddr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
